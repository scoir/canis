/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/protogen/common"
)

var invitationGetCmd = &cobra.Command{
	Use:   "get AGENT_NAME",
	Short: "Generate a connection invitation to the specified agent.",
	RunE:  invitationGet,
	Args:  cobra.ExactArgs(1),
}

func init() {
	invitationsCmd.AddCommand(invitationGetCmd)
	invitationGetCmd.Flags().StringVar(&subject, "subject", "", "subject name for the destination of the invitation")
	_ = invitationGetCmd.MarkFlagRequired("subject")
}

func invitationGet(_ *cobra.Command, args []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	ctx := context.Background()

	req := &common.InvitationRequest{
		AgentName:  args[0],
		ExternalId: subject,
	}
	invite, err := cli.GetAgentInvitation(ctx, req)
	if err != nil {
		return errors.Wrap(err, "unable to get invitation")
	}

	fmt.Println(invite.Invitation)
	return nil
}
