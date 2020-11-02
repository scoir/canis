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

	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/framework"
)

var (
	cfgFile        string
	prov           *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm load balancer.",
	Long: `"The canis didcomm load balancer but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	lock                 secretlock.Service
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
		DefaultConfigName: "canis-lb-config",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-lb-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	conf := configProvider.Load(cfgFile).
		WithLedgerStore().
		WithAMQP().
		WithMasterLockKey()

	lc, err := conf.LedgerStore()
	if err != nil {
		log.Fatalln("invalid ledgerstore key in configuration")
	}

	ls, err := lc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	lock, err := local.NewService(strings.NewReader(conf.MasterLockKey()), nil)
	if err != nil {
		log.Fatalln("unable to create lock service", err)
	}

	prov = &Provider{
		conf:                 conf,
		lock:                 lock,
		ariesStorageProvider: ls,
	}
}

func (r *Provider) GetAMQPAddress() string {
	return r.conf.AMQPAddress()
}

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("loadbalancer.grpc")
}

func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	return nil, errors.New("not supported")
}
