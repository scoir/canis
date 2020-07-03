/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/storage/leveldb"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/framework"
)

type Config struct {
	ctx *context.Provider

	*framework.GRPCConfig
	*framework.AgentConfig
	*framework.DatastoreConfig
}

func (r *Config) GetAriesContext() *context.Provider {
	if r.ctx == nil {
		err := r.createAriesContext()
		if err != nil {
			log.Fatalln("failed to create aries context", err)
		}

	}
	return r.ctx
}

func (r *Config) createAriesContext() error {
	wsinbound := r.WSInbound.Address()

	log.Printf("creating for %s\n", r.DBPath)
	fw, err := aries.New(
		aries.WithMessageServiceProvider(msghandler.NewRegistrar()),
		aries.WithStoreProvider(leveldb.NewProvider(r.DBPath)),
		defaults.WithInboundWSAddr(wsinbound, wsinbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithServiceEndpoint(r.Address()),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to start aries agent rest on port [%s], failed to initialize framework",
			wsinbound)
	}

	ctx, err := fw.Context()
	if err != nil {
		return errors.Wrapf(err, "failed to start aries agent rest on port [%s], failed to get aries context",
			wsinbound)
	}
	r.ctx = ctx

	return nil
}
