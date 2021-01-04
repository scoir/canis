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

var agentsDeleteCmd = &cobra.Command{
	Use:       "delete AGENT_NAME",
	Short:     "Generate a connection invitation to the specified agent.",
	RunE:      agentsDelete,
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"AGENT_NAME"},
}

func init() {
	agentsCmd.AddCommand(agentsDeleteCmd)
	agentsDeleteCmd.Flags().StringVar(&subject, "subject", "", "subject name for the destination of the invitation")
	_ = agentsDeleteCmd.MarkFlagRequired("subject")
}

func agentsDelete(_ *cobra.Command, args []string) error {
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
