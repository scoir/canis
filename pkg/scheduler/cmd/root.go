/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/client/canis"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
	"github.com/scoir/canis/pkg/runtime"
	"github.com/scoir/canis/pkg/runtime/docker"
)

const (
	executionKey = "execution"
)

var cfgFile string

var ctx *Provider

var rootCmd = &cobra.Command{
	Use:   "canis-scheduler",
	Short: "The canis orchestration service.",
	Long: `"The canis orchestration service.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp   *viper.Viper
	exec runtime.Executor
	dm   *manager.DataProviderManager
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-scheduler-config.yaml)")
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
		vp.SetConfigName("canis-scheduler-config")
	}

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

func (r *Provider) Executor() (runtime.Executor, error) {
	if r.exec != nil {
		return r.exec, nil
	}

	rtc := &context.RuntimeConfig{}
	err := r.vp.UnmarshalKey(executionKey, rtc)
	if err != nil {
		return nil, errors.Wrap(err, "execution environment is not correctly configured")
	}
	switch rtc.Runtime {
	case "kubernetes":
		r.exec, err = r.loadK8s()
	case "docker":
		r.exec, err = r.loadDocker(rtc.Docker)
	default:
		return nil, errors.New("no known execution environment is configured")
	}

	return r.exec, errors.Wrap(err, "unable to launch runtime from config")

}

func (r *Provider) loadK8s() (runtime.Executor, error) {
	return nil, errors.New("not implemented")
}

func (r *Provider) loadDocker(dc *docker.Config) (runtime.Executor, error) {
	if dc == nil {
		return nil, errors.New("docker execution environment not properly configured")
	}
	d, err := docker.New(dc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create docker execution environment")
	}
	return d, nil
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

func (r *Provider) CanisClient() (*canis.Client, error) {
	if !r.vp.IsSet("api.grpc") {
		return nil, errors.New("api client is not properly configured")
	}

	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "api client is not properly configured")
	}

	client := canis.New(ep.Address())
	return client, nil
}
