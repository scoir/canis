/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/apiserver"
	"github.com/scoir/canis/pkg/controller"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the steward orchestration service",
	Long:  `Starts a steward orchestration service`,
	Run:   runStart,
}

func runStart(cmd *cobra.Command, args []string) {

	srv, err := apiserver.New(ctx)
	if err != nil {
		log.Fatalln("error initializing canis-apiserver", err)
	}

	runner, err := controller.New(ctx, srv)

	if err != nil {
		log.Fatalln("unable to start canis-apiserver", err)
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
