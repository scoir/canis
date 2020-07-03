/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/controller"
	"github.com/scoir/canis/pkg/schema"
	"github.com/scoir/canis/pkg/steward"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the steward orchestration service",
	Long:  `Starts a steward orchestration service`,
	Run:   runStart,
}

func runStart(cmd *cobra.Command, args []string) {
	ctx := config.GetAriesContext()

	client := schema.New()
	agent, err := steward.New(ctx, config, client)
	if err != nil {
		log.Fatalln("error initializing steward agent", err)
	}

	runner, err := controller.New(
		ctx,
		config.GRPC.Host,
		config.GRPC.Port,
		config.GRPCBridge.Host,
		config.GRPCBridge.Port,
		agent)

	if err != nil {
		log.Fatalln("unable to start steward", err)
	}

	err = runner.Launch()
	if err != nil {
		log.Fatalln("launch errored with", err)
	}

	log.Println("Shutdown")
}

func init() {
	rootCmd.AddCommand(startCmd)
}
