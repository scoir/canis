/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
	couchdbstore "github.com/hyperledger/aries-framework-go/pkg/storage/couchdb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/framework"
)

var (
	cfgFile string
	ctx     *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm service.",
	Long: `"The canis didcomm service but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp *viper.Viper
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
		vp.SetConfigName("canis-didcomm")
	}

	vp.SetEnvPrefix("CANIS")
	vp.AutomaticEnv() // read in environment variables that match
	_ = vp.BindPFlags(pflag.CommandLine)

	// If a vp file is found, read it in.
	if err := vp.ReadInConfig(); err != nil {
		fmt.Println("unable to read vp:", vp.ConfigFileUsed(), err)
		os.Exit(1)
	}

	ctx = &Provider{
		vp: vp,
	}
}

// GetStorageProvider todo
func (r *Provider) GetStorageProvider() (storage.Provider, error) {
	store := r.vp.GetString("protocol.store")
	if store != "couchdb" {
		return nil, fmt.Errorf("unsupported store: %s", store)
	}

	url := r.vp.GetString("protocol.url")
	cdb, err := couchdbstore.NewProvider(fmt.Sprintf("http://%s:5984/", url))
	if err != nil {
		return nil, err
	}

	return cdb, nil
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
