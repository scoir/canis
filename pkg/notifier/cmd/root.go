/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/amqp"
	"github.com/scoir/canis/pkg/amqp/rabbitmq"
	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/datastore"
)

var (
	cfgFile        string
	prov           *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-notifier",
	Short: "The canis webhook notifier.",
	Long: `"The canis webhook notifier.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	store datastore.Store
	conf  config.Config
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
		DefaultConfigName: "canis-webhook-notifier",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-webhook-notifier.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	conf := configProvider.Load(cfgFile).
		WithDatastore().
		WithAMQP()

	dc, err := conf.DataStore()
	if err != nil {
		log.Fatalln("invalid datastore key in configuration")
	}

	sp, err := dc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := sp.Open()
	if err != nil {
		log.Fatalln("unable to open datastore")
	}

	prov = &Provider{
		conf:  conf,
		store: store,
	}
}

func (r *Provider) GetDatastore() datastore.Store {
	return r.store
}

func (r *Provider) GetAMQPListener(queue string) amqp.Listener {
	conf, err := r.conf.AMQPConfig()
	if err != nil {
		log.Fatalln("unexpected error reading amqp config", err)
	}

	l, err := rabbitmq.NewListener(conf.Endpoint(), queue)
	if err != nil {
		log.Fatalln("unable to intialize new amqp listener", err)
	}

	return l
}
