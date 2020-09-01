/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/controller"
	"github.com/scoir/canis/pkg/didcomm"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm service",
	Long:  `Starts a didcomm service`,
	Run:   runStart,
}

func runStart(cmd *cobra.Command, args []string) {

	srv, err := didcomm.New(ctx)
	if err != nil {
		log.Fatalln("error initializing canis-didcomm", err)
	}

	runner, err := controller.New(ctx, srv)

	if err != nil {
		log.Fatalln("unable to start canis-didcomm", err)
	}

	err = runner.Launch()
	if err != nil {
		log.Fatalln("launch error", err)
	}

	log.Println("Shutdown")
}

func init() {
	rootCmd.AddCommand(startCmd)
}
