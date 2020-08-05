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

var seed string

// stewardInitCmd represents the init command
var stewardInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the steward's public key in a canis credential hub",
	Long:  `Initialize the steward's public key in a canis credential hub`,
	Run:   initCluster,
}

func init() {
	stewardCmd.AddCommand(stewardInitCmd)
	stewardInitCmd.Flags().StringVar(&seed, "seed", "", "seed for public DID")
}

func initCluster(_ *cobra.Command, _ []string) {
	if seed == "" {
		log.Fatalln("seed flag is required")
	}

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

	logs, err := executor.InitSteward(seed, d)
	if err != nil {
		log.Println("error initializing steward", err)
	}

	fmt.Println(logs)
}
