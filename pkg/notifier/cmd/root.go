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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
)

var (
	cfgFile string
	prov    *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-notifier",
	Short: "The canis webhook notifier.",
	Long: `"The canis webhook notifier.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp    *viper.Viper
	store datastore.Store
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-webhook-notifier.yaml)")
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
		vp.SetConfigName("canis-webhook-notifier")
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

	sp, err := dc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := sp.Open()
	if err != nil {
		log.Fatalln("unable to open datastore")
	}

	prov = &Provider{
		vp:    vp,
		store: store,
	}
}

func (r *Provider) GetDatastore() datastore.Store {
	return r.store
}

func (r *Provider) GetAMQPAddress() string {
	amqpUser := r.vp.GetString("amqp.user")
	amqpPwd := r.vp.GetString("amqp.password")
	amqpHost := r.vp.GetString("amqp.host")
	amqpPort := r.vp.GetInt("amqp.port")
	amqpVHost := r.vp.GetString("amqp.vhost")
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", amqpUser, amqpPwd, amqpHost, amqpPort, amqpVHost)
}
