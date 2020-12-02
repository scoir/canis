/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/framework/context"
)

var (
	cfgFile        string
	ctx            *context.Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "sirius",
	Short: "The canis CLI controls the Canis Credential Hub.",
	Long: `The canis CLI controls the Canis Credential Hub.

 Find more information at: https://canis.network/docs/reference/canis/overview`,
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
		DefaultConfigName: ".canis",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "no default config file")
}

func initConfig() {
	ctx = context.NewProvider(configProvider.Load(cfgFile))
}
