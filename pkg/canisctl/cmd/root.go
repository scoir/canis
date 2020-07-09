/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/framework/context"
)

var cfgFile string

var config *viper.Viper
var ctx *context.Provider

var rootCmd = &cobra.Command{
	Use:   "canisctl",
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.canis.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config = viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		config.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".canis" (without extension).
		config.AddConfigPath(home)
		config.SetConfigName(".canis")
	}

	config.AutomaticEnv() // read in environment variables that match
	_ = config.BindPFlags(pflag.CommandLine)

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		fmt.Println("unable to read config:", config.ConfigFileUsed(), err)
		os.Exit(1)
	}

	ctx = context.NewProvider(config)
}
