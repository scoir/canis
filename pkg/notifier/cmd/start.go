/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/notifier"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the webhook notifier",
	Long:  `Starts a webhook notifier`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	log.Println("starting webhook notifier")

	srv, err := notifier.New(prov)
	if err != nil {
		log.Fatalln("unable to launch webhook notifier", err)
	}

	err = srv.Start()
	if err != nil {
		log.Println("webhook notifier exited with error", err)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
