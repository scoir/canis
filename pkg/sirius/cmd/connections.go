/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/spf13/cobra"
)

var connectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "Manage connections for running agents",
	Long: `Manage connections for running agents

 Find more information at: https://canis.network/docs/reference/canis/overview`,
}

func init() {
	rootCmd.AddCommand(connectionsCmd)
}
