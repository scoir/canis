/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/spf13/cobra"
)

var stewardCmd = &cobra.Command{
	Use:   "steward",
	Short: "Commands for initializing and controlling the steward.",
	Long:  `Commands for initializing and controlling the steward.`,
}

func init() {
	rootCmd.AddCommand(stewardCmd)
}
