/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/spf13/cobra"
)

var invitationsCmd = &cobra.Command{
	Use:   "invitations",
	Short: "Manage connection invitations for running agents",
}

func init() {
	connectionsCmd.AddCommand(invitationsCmd)
}
