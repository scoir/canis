/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package indy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"

	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

type indyClient interface {
	GetNym(did string) (*vdr.ReadReply, error)
	GetEndpoint(did string) (*vdr.ReadReply, error)
	RefreshPool() error
	Close() error
}

//Implements the VDRI interface for Hyperledger Indy networks
type VDRI struct {
	methodName string
	refresh    bool
	client     indyClient
}

func New(methodName string, opts ...Option) (*VDRI, error) {
	vdri := &VDRI{methodName: methodName}

	for _, opt := range opts {
		opt(vdri)
	}

	if vdri.client == nil {
		return nil, errors.New("an Indy Ledger client must be set with an option to New")
	}

	err := vdri.client.RefreshPool()
	if err != nil {
		return nil, fmt.Errorf("refreshing indy pool failed: %w", err)
	}

	return vdri, nil
}

func (r *VDRI) Store(doc *did.Doc, by *[]vdriapi.ModifiedBy) error {
	return nil
}

func (r *VDRI) Accept(method string) bool {
	return method == r.methodName
}

func (r *VDRI) Close() error {
	return r.client.Close()
}

// Option configures the peer vdri
type Option func(opts *VDRI)

// WithAccept option is for accept did method
func WithRefresh(refresh bool) Option {
	return func(opts *VDRI) {
		opts.refresh = refresh
	}
}

func WithIndyClient(client indyClient) Option {
	return func(opts *VDRI) {
		opts.client = client
	}
}

func WithIndyVDRGenesisFile(genesisFile string) Option {
	return func(opts *VDRI) {
		gfr, err := os.Open(genesisFile)
		if err != nil {
			log.Println("unable to open genesis file", err)
			return
		}

		opts.client, err = vdr.New(gfr)
		if err != nil {
			err = fmt.Errorf("error connecting to indy ledger: (%w)", err)
			log.Println(err)
		}

	}
}

func WithIndyVDRGenesisReader(genesisData io.ReadCloser) Option {
	return func(opts *VDRI) {
		var err error
		opts.client, err = vdr.New(genesisData)
		if err != nil {
			err = fmt.Errorf("error connecting to indy ledger: (%w)", err)
			log.Println(err)
		}

	}
}
