/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/amqp"
	"github.com/scoir/canis/pkg/amqp/rabbitmq"
	transportamqp "github.com/scoir/canis/pkg/aries/transport/amqp"
	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
)

var (
	cfgFile        string
	ctx            *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm doorman service.",
	Long: `"The canis didcomm doorman service but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	store                datastore.Store
	ariesStorageProvider storage.Provider
	conf                 config.Config
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	configProvider = &config.ViperConfigProvider{
		DefaultConfigName: "canis-doorman-config",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-doorman-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	conf := configProvider.Load(cfgFile).
		WithDatastore().
		WithLedgerStore().
		WithAMQP().
		WithVDRI().
		WithMasterLockKey()

	dc, err := conf.DataStore()
	if err != nil {
		log.Fatalln("invalid datastore key in configuration", err)
	}

	sp, err := dc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := sp.Open()
	if err != nil {
		log.Fatalln("unable to open datastore")
	}

	lc, err := conf.LedgerStore()
	if err != nil {
		log.Fatalln("invalid ledgerstore key in configuration")
	}

	ls, err := lc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	ctx = &Provider{
		store:                store,
		ariesStorageProvider: ls,
		conf:                 conf,
	}
}

// GetAriesStorageProvider todo
func (r *Provider) GetAriesStorageProvider() storage.Provider {
	return r.ariesStorageProvider
}

func (r *Provider) GetDatastore() (datastore.Store, error) {
	return r.store, nil
}

// GetGRPCEndpoint todo
func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("api.grpc")
}

// GetBridgeEndpoint todo
func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("api.grpcBridge")
}

func (r *Provider) GetAMQPPublisher(queue string) amqp.Publisher {
	cfg, err := r.conf.AMQPConfig()
	if err != nil {
		log.Fatalln("unexpected error reading amqp config", err)
	}

	pub, err := rabbitmq.NewPublisher(cfg.Endpoint(), queue)
	if err != nil {
		log.Fatalln("unable to launch rabbitmq publisher", err)
	}

	return pub
}

func (r *Provider) GetAriesContext() (*ariescontext.Provider, error) {
	external := r.conf.GetString("inbound.external")
	cfg, err := r.conf.AMQPConfig()
	if err != nil {
		return nil, err
	}

	lock, err := local.NewService(strings.NewReader(r.conf.MasterLockKey()), nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create lock service")
	}

	vdrisConfig, err := r.conf.VDRIs()
	if err != nil {
		return nil, err
	}

	vdris, err := context.GetAriesVDRIs(vdrisConfig)
	if err != nil {
		return nil, err
	}

	amqpInbound, err := transportamqp.NewInbound(cfg.Endpoint(), external, "didexchange", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create amqp inbound")
	}
	vopts := []aries.Option{
		aries.WithStoreProvider(r.ariesStorageProvider),
		aries.WithInboundTransport(amqpInbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithSecretLock(lock),
	}
	for _, vdri := range vdris {
		vopts = append(vopts, aries.WithVDRI(vdri))
	}

	ar, err := aries.New(vopts...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create aries defaults")
	}

	actx, err := ar.Context()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries context")
	}

	return actx, err
}
