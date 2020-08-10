/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
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
	"github.com/scoir/canis/pkg/datastore/mongodb"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/runtime"
	"github.com/scoir/canis/pkg/runtime/docker"
)

const (
	executionKey = "execution"
)

var cfgFile string

var ctx *Provider

var rootCmd = &cobra.Command{
	Use:   "steward",
	Short: "The canis steward orchestration service.",
	Long: `"The canis steward orchestration service.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp     *viper.Viper
	exec   runtime.Executor
	ds     datastore.Provider
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

	vp.AutomaticEnv() // read in environment variables that match
	_ = vp.BindPFlags(pflag.CommandLine)

	// If a vp file is found, read it in.
	if err := vp.ReadInConfig(); err != nil {
		fmt.Println("unable to read vp:", vp.ConfigFileUsed(), err)
		os.Exit(1)
	}

	ctx = &Provider{vp: vp}
}

func (r *Provider) StorageProvider() (datastore.Provider, error) {
	if r.ds != nil {
		return r.ds, nil
	}

	var err error
	r.ds, err = mongodb.NewProvider(&mongodb.Config{
		URL:      r.vp.GetString("Datastore.Mongo.URL"),
		Database: r.vp.GetString("Datastore.Mongo.Database"),
	})
	if err != nil {
		log.Fatalln("unable to create datastore connection")
	}

	return r.ds, nil
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

func (r *Provider) VDR() (*vdr.Client, error) {
	genesisFile := r.vp.GetString("genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	return cl, nil
}

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	return nil, nil
}
func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	return nil, nil
}
