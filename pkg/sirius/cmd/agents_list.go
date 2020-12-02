/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	api "github.com/scoir/canis/pkg/apiserver/api/protogen"
)

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Accept a connection invitation from another agent.",
	Args:  cobra.ExactArgs(0),
	RunE:  agentsList,
}

func init() {
	agentsCmd.AddCommand(agentsListCmd)
}

func agentsList(cmd *cobra.Command, _ []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	ctx := context.Background()

	req := &api.ListAgentRequest{
		Name: "",
	}

	agents, err := cli.ListAgent(ctx, req)
	if err != nil {
		return errors.Wrap(err, "unable to get invitation")
	}

	tab := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	cmd.SetOut(tab)

	cmd.Print(strings.Join([]string{"NAME", "STATUS", "PUBLIC DID"}, "\t"), "\n")
	for _, agent := range agents.Agents {
		statusName := api.Agent_Status_name[int32(agent.Status)]

		cmd.Print(strings.Join([]string{agent.Name, statusName, strconv.FormatBool(agent.PublicDid)}, "\t"), "\n")
	}

	err = tab.Flush()
	return err
}
