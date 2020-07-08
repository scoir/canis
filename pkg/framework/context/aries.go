/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"log"
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/route"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms/legacykms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage/leveldb"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/schema"
)

const (
	dbPathKey    = "dbpath"
	wsinboundKey = "wsinbound"
)

type AgentConfig struct {
	framework.Endpoint
	DBPath     string             `mapstructure:"dbpath"`
	WSInbound  framework.Endpoint `mapstructure:"wsinbound"`
	GRPC       framework.Endpoint `mapstructure:"grpc"`
	GRPCBridge framework.Endpoint `mapstructure:"grpcbridge"`
	LedgerURL  string             `mapstructure:"ledgerURL"`

	GetAriesOptions func() []aries.Option

	lock sync.Mutex
}

type provider struct {
	sp  storage.Provider
	kms *legacykms.BaseKMS
}

func newProvider(dbpath string) *provider {
	r := &provider{
		sp: leveldb.NewProvider(dbpath),
	}

	r.kms, _ = legacykms.New(r)
	return r
}

func (r *provider) StorageProvider() storage.Provider {
	return r.sp
}

func (r *provider) createKMS(_ api.Provider) (api.CloseableKMS, error) {
	return r.kms, nil
}

func (r *Provider) GetAriesContext() *context.Provider {
	if r.ctx == nil {
		err := r.createAriesContext()
		if err != nil {
			log.Fatalln("failed to create aries context", err)
		}

	}
	return r.ctx
}

func (r *Provider) createAriesContext() error {
	framework, err := aries.New(r.getOptions()...)
	if err != nil {
		return errors.Wrap(err, "failed to start aries agent rest, failed to initialize framework")
	}

	ctx, err := framework.Context()
	if err != nil {
		return errors.Wrap(err, "failed to start aries agent rest on port, failed to get aries context")
	}
	r.ctx = ctx

	return nil
}

func (r *Provider) getOptions() []aries.Option {
	var out []aries.Option

	dbpath := r.vp.GetString(dbPathKey)
	if dbpath != "" {
		out = append(out, aries.WithStoreProvider(leveldb.NewProvider(dbpath)))
	} else {
		out = append(out, aries.WithStoreProvider(mem.NewProvider()))
	}

	if r.vp.IsSet(wsinboundKey) {
		wsinbound := &framework.Endpoint{}
		_ = r.vp.UnmarshalKey(wsinboundKey, wsinbound)
		out = append(out, defaults.WithInboundWSAddr(wsinbound.Address(), wsinbound.Address()))
	}

	if r.vp.IsSet("host") && r.vp.IsSet("port") {
		ep := &framework.Endpoint{}
		_ = r.vp.Unmarshal(ep)
		out = append(out, aries.WithServiceEndpoint(ep.Address()))
	}

	//TODO:  do we need configuration options to turn on or off the following 2 options?
	out = append(out, []aries.Option{
		aries.WithMessageServiceProvider(msghandler.NewRegistrar()),
		aries.WithOutboundTransports(ws.NewOutbound()),
	}...)

	return out
}

func (r *Provider) GetDIDClient() (*didexchange.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.didcl != nil {
		return r.didcl, nil
	}

	didcl, err := didexchange.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "error creating did client")
	}

	r.didcl = didcl
	return r.didcl, nil
}

func (r *Provider) GetSchemaClient() (*schema.Client, error) {
	return schema.New(), nil
}

func (r *Provider) GetCredentialClient() (*issuecredential.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.credcl != nil {
		return r.credcl, nil
	}

	credcl, err := issuecredential.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "error creating credential client")
	}
	r.credcl = credcl
	return r.credcl, nil
}

func (r *Provider) GetRouterClient() (*route.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.routecl != nil {
		return r.routecl, nil
	}

	routecl, err := route.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create route client for college: %v\n")
	}
	r.routecl = routecl
	return r.routecl, nil
}
