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

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/framework"
	indywrapper "github.com/scoir/canis/pkg/indy"
)

var (
	cfgFile        string
	prov           *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis resolver.",
	Long: `"The canis  resolver but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	indyClient indywrapper.IndyVDRClient
	conf       config.Config
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
		DefaultConfigName: "http-indy-resolver",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/http-indy-resolver.yml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	conf := configProvider.Load(cfgFile).
		WithLedgerGenesis()

	genesisFile := conf.LedgerGenesis()
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		log.Fatalln("unable to create VDR client")
	}

	prov = &Provider{
		conf:       conf,
		indyClient: cl,
	}

}

func (r *Provider) GetHTTPEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("resolver.http")
}

func (r *Provider) IndyVDR() indywrapper.IndyVDRClient {
	return r.indyClient
}
