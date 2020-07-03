/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/scoir/canis/pkg/agent"
)

var cfgFile string

var config *agent.Config

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "The canis agent credential issuer/resolver.",
	Long: `"The canis agent credential issuer/resolver.

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.canis.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		cfgFile = strings.Join([]string{home, ".canis"}, string(os.PathSeparator))
	}

	// If a config file is found, read it in.
	f, err := os.Open(cfgFile)
	if err != nil {
		fmt.Println("unable to read config:", cfgFile, err)
		os.Exit(1)
	}

	config = &agent.Config{}
	err = yaml.NewDecoder(f).Decode(config)
	if err != nil {
		fmt.Println("failed to unmarshal config", err)
		os.Exit(1)
	}

}
