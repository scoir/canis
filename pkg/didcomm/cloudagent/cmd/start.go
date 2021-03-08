/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/didcomm/cloudagent"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm cloudagent service",
	Long:  `Starts a didcomm cloudagent service`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	i, err := cloudagent.New(ctx)
	if err != nil {
		log.Fatalln("unable to initialize mediator", err)
	}

	err = i.Start()
	if err != nil {
		log.Println("cloudagent exited with", err)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
