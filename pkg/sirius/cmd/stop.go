/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// stopCmd represents the start command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Starts a canis credential hub",
	Long:  `Starts a canis credential hub.`,
	Run:   stopCluster,
}

func stopCluster(_ *cobra.Command, _ []string) {
	executor, err := ctx.Executor()
	if err != nil {
		log.Fatalln("unable to access executor", err)
	}

	err = executor.ShutdownSteward()
	if err != nil {
		log.Println("error stopping steward", err)
	}

	fmt.Println("cluster is stopped")
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
