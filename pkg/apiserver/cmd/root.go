/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/runtime"
)

var (
	cfgFile string
	ctx     *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-apiserver",
	Short: "The canis steward orchestration service.",
	Long: `"The canis steward orchestration service.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp     *viper.Viper
	exec   runtime.Executor
	dm     *manager.DataProviderManager
	client *vdr.Client
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-apiserver-config.yaml)")
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
		vp.AddConfigPath("./config/docker/")
		vp.SetConfigName("canis-apiserver-config")
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
	ctx = &Provider{
		vp: vp,
		dm: dm,
	}
}

func (r *Provider) StorageManager() *manager.DataProviderManager {
	return r.dm
}

func (r *Provider) StorageProvider() (datastore.Provider, error) {
	return r.dm.DefaultStoreProvider()
}

func (r *Provider) VDR() (*vdr.Client, error) {
	genesisFile := r.vp.GetString("genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	status, err := cl.GetPoolStatus()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get pool status")
	}

	d, _ := json.MarshalIndent(status, " ", " ")
	fmt.Println(string(d))

	return cl, nil
}

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpcBridge", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc bridge is not properly configured")
	}

	return ep, nil
}