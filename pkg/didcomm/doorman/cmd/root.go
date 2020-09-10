/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/aries/transport/amqp"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/framework"
)

var (
	cfgFile string
	ctx     *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm doorman service.",
	Long: `"The canis didcomm doorman service but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp                   *viper.Viper
	store                datastore.Store
	ariesStorageProvider storage.Provider
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-didcomm-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	vp := viper.New()
	if cfgFile != "" {
		// Use vp file from the flag.
		vp.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		vp.SetConfigType("yaml")
		vp.AddConfigPath("/etc/canis/")
		vp.AddConfigPath("./deploy/compose/")
		vp.SetConfigName("canis-didcomm-config")
	}

	vp.SetEnvPrefix("CANIS")
	vp.AutomaticEnv() // read in environment variables that match
	_ = vp.BindPFlags(pflag.CommandLine)

	// If a vp file is found, read it in.
	if err := vp.ReadInConfig(); err != nil {
		fmt.Println("unable to read vp:", vp.ConfigFileUsed(), err)
		os.Exit(1)
	}

	dc := &framework.DatastoreConfig{}
	err := vp.UnmarshalKey("datastore", dc)
	if err != nil {
		log.Fatalln("invalid datastore key in configuration")
	}

	dm := manager.NewDataProviderManager(dc)
	sp, err := dm.DefaultStoreProvider()
	if err != nil {
		log.Fatalln("unable to retrieve default storage provider", err)
	}

	store, err := sp.OpenStore("canis")
	if err != nil {
		log.Fatalln(err, "unable to open canis data store")
	}

	asp, err := store.GetAriesProvider()
	if err != nil {
		log.Fatalln(err, "unable to open aries storage provider")
	}

	ctx = &Provider{
		vp:                   vp,
		store:                store,
		ariesStorageProvider: asp,
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
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc is not properly configured")
	}

	return ep, nil
}

// GetBridgeEndpoint todo
func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpcBridge", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc bridge is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetAriesContext() (*ariescontext.Provider, error) {
	external := r.vp.GetString("inbound.external")
	config := &framework.AMQPConfig{}
	err := r.vp.UnmarshalKey("inbound.amqp", config)

	amqpInbound, err := amqp.NewInbound(config.Endpoint(), external, "didexchange", "", "")
	ar, err := aries.New(
		aries.WithStoreProvider(r.ariesStorageProvider),
		aries.WithInboundTransport(amqpInbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create aries defaults")
	}

	actx, err := ar.Context()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries context")
	}

	return actx, err
}
