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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/framework"
)

var (
	cfgFile string
	prov    *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm load balancer.",
	Long: `"The canis didcomm load balancer but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp                   *viper.Viper
	lock                 secretlock.Service
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-lb-config.yaml)")
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
		vp.SetConfigName("canis-lb-config")
	}

	vp.SetEnvPrefix("CANIS")
	vp.AutomaticEnv() // read in environment variables that match
	_ = vp.BindPFlags(pflag.CommandLine)

	// If a vp file is found, read it in.
	if err := vp.ReadInConfig(); err != nil {
		fmt.Println("unable to read vp:", vp.ConfigFileUsed(), err)
		os.Exit(1)
	}

	lc := &framework.LedgerStoreConfig{}
	err := vp.UnmarshalKey("ledgerstore", lc)
	if err != nil {
		log.Fatalln("invalid ledgerstore key in configuration")
	}

	ls, err := lc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	mlk := vp.GetString("masterLockKey")
	if mlk == "" {
		mlk = "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ="
	}

	lock, err := local.NewService(strings.NewReader(mlk), nil)
	if err != nil {
		log.Fatalln("unable to create lock service", err)
	}

	prov = &Provider{
		vp:                   vp,
		lock:                 lock,
		ariesStorageProvider: ls,
	}
}

func (r *Provider) GetAMQPAddress() string {
	amqpUser := r.vp.GetString("amqp.user")
	amqpPwd := r.vp.GetString("amqp.password")
	amqpHost := r.vp.GetString("amqp.host")
	amqpPort := r.vp.GetInt("amqp.port")
	amqpVHost := r.vp.GetString("amqp.vhost")
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", amqpUser, amqpPwd, amqpHost, amqpPort, amqpVHost)
}

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("loadbalancer.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	return nil, errors.New("not supported")
}
