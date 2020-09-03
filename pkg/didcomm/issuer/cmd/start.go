/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/controller"
	"github.com/scoir/canis/pkg/didcomm/issuer"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm service",
	Long:  `Starts a didcomm service`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	i, err := issuer.New(ctx)
	if err != nil {
		log.Fatalln("unable to initialize issuer", err)
	}

	runner, err := controller.New(ctx, i)
	if err != nil {
		log.Fatalln("unable to start didcomm-issuer", err)
	}

	err = runner.Launch()
	if err != nil {
		log.Fatalln("launch errored with", err)
	}

}

func init() {
	rootCmd.AddCommand(startCmd)
}
