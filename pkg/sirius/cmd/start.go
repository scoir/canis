/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a canis credential hub",
	Long:  `Starts a canis credential hub.`,
	Run:   startCluster,
}

func startCluster(_ *cobra.Command, _ []string) {
	executor, err := ctx.Executor()
	if err != nil {
		log.Fatalln("unable to access executor", err)
	}

	out, err := ctx.GetStewardConfig()
	if err != nil {
		log.Fatalln("unable to generate steward config", err)
	}
	d, err := yaml.Marshal(out)
	if err != nil {
		log.Fatalln("unable to marshal steward config")
	}

	pid, err := executor.LaunchSteward(d)
	if err != nil {
		log.Println("error launching steward", err)
	}

	fmt.Println("steward launched at", pid)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
