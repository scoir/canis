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

	api "github.com/scoir/canis/pkg/apiserver/api/protogen"
)

var schemaNames []string
var publicDID bool

var agentsCreateCmd = &cobra.Command{
	Use:   "create AGENT_NAME",
	Short: "Generate a connection invitation to the specified agent.",
	RunE:  agentsCreate,
	Args:  cobra.ExactArgs(1),
}

func init() {
	agentsCmd.AddCommand(agentsCreateCmd)
	agentsCreateCmd.Flags().StringArrayVar(&schemaNames, "schema-name", []string{}, "list of schema this agent is allowed to issue")
	agentsCreateCmd.Flags().BoolVar(&publicDID, "public-did", false, "assign a public DID to this agent if flag is set")
}

func agentsCreate(_ *cobra.Command, args []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	ctx := context.Background()

	agentName := args[0]
	req := &api.CreateAgentRequest{
		Agent: &api.NewAgent{
			Name:                  agentName,
			EndorsableSchemaNames: schemaNames,
			PublicDid:             publicDID,
		},
	}

	_, err = cli.CreateAgent(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "unable to create agent %s", agentName)
	}

	fmt.Printf("AGENT %s CREATED\n", agentName)
	return nil
}
