/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ariesctx "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	ndid "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/steward/api"
)

type Agent struct {
	self      *Self
	steward   api.AdminClient
	bouncer   ndid.Bouncer
	persister persistence
}

type provider interface {
	UnmarshalConfig(dest interface{}) error
	Datastore() (datastore.Provider, error)
	GetStewardClient() (api.AdminClient, error)
	GetDIDClient() (*didexchange.Client, error)
	GetAriesContext() *ariesctx.Provider
}

type Option func(opts *Self)

func NewAgent(p provider, opts ...Option) (*Agent, error) {
	r := &Agent{self: &Self{}}

	for _, opt := range opts {
		opt(r.self)
	}

	d, _ := json.MarshalIndent(r.self, " ", " ")
	fmt.Println(string(d))

	ds, err := p.Datastore()
	if err != nil {
		return nil, errors.Wrap(err, "agent can not load datastore")
	}

	r.persister, err = newPersistence(ds)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent persistence")
	}

	r.steward, err = p.GetStewardClient()
	if err != nil {
		return nil, errors.Wrap(err, "error getting steward client for agent")
	}

	r.bouncer, err = ndid.NewBouncer(p)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create did bouncer for agent")
	}

	fmt.Println("getting self")
	self, err := r.persister.GetSelf(r.self.ID)
	if err != nil {
		self = &Self{}
		if r.self.HasPublicDID {
			aries := p.GetAriesContext()
			did, err := aries.VDRIRegistry().Create("scr")
			if err != nil {
				return nil, errors.Wrap(err, "unable to bootstrap agent")
			}
			self.PublicDID = did.ID
		}
		err = r.persister.SaveSelf(self)
		if err != nil {
			return nil, errors.Wrap(err, "unable to save agent identity")
		}
	}

	r.self = self

	return r, nil
}

// WithPublicDID option is for accept did method
func WithPublicDID(pd bool) Option {
	return func(opts *Self) {
		opts.HasPublicDID = pd
	}
}

// WithAgentID option is for accept did method
func WithAgentID(id string) Option {
	return func(opts *Self) {
		opts.ID = id
	}
}
