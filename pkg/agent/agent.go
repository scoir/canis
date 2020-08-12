/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	ariesctx "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/datastore"
	ndid "github.com/scoir/canis/pkg/didexchange"
)

type Agent struct {
	ID           string
	Endpoint     string
	HasPublicDID bool
	PublicDID    string
	steward      api.AdminClient
	bouncer      ndid.Bouncer
}

type provider interface {
	UnmarshalConfig(dest interface{}) error
	Datastore() (datastore.Provider, error)
	GetAPIAdminClient() (api.AdminClient, error)
	GetDIDClient() (*didexchange.Client, error)
	GetAriesContext() *ariesctx.Provider
}

type Option func(opts *Agent)

func NewAgent(p provider, opts ...Option) (*Agent, error) {
	r := &Agent{}

	for _, opt := range opts {
		opt(r)
	}

	var err error
	r.steward, err = p.GetAPIAdminClient()
	if err != nil {
		return nil, errors.Wrap(err, "error getting steward client for agent")
	}

	r.bouncer, err = ndid.NewBouncer(p)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create did bouncer for agent")
	}

	if r.HasPublicDID {
		aries := p.GetAriesContext()
		did, err := aries.VDRIRegistry().Create("scr", vdri.WithServiceEndpoint(r.Endpoint))
		if err != nil {
			return nil, errors.Wrap(err, "unable to bootstrap agent")
		}
		r.PublicDID = did.ID
	}
	return r, nil
}

// WithPublicDID option is for accept did method
func WithPublicDID(pd bool) Option {
	return func(opts *Agent) {
		opts.HasPublicDID = pd
	}
}

// WithAgentID option is for accept did method
func WithAgentID(id string) Option {
	return func(opts *Agent) {
		opts.ID = id
	}
}
