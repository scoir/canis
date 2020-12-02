/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/spf13/cobra"
)

var credentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage credentials for running agents",
}

func init() {
	rootCmd.AddCommand(credentialsCmd)
}
