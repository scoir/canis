/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/scheduler"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the canis orchestration service",
	Long:  `Starts the canis orchestration service`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {

	srv, err := scheduler.New(ctx)
	if err != nil {
		log.Fatalln("error initializing canis-apiserver", err)
	}

	srv.Run()
	log.Println("Shutdown")
}

func init() {
	rootCmd.AddCommand(startCmd)
}
