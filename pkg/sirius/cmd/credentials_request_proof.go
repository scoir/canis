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

var credentialsRequestProofCmd = &cobra.Command{
	Use:   "request-proof AGENT_NAME",
	Short: "Request a proof presentation on behalf of AGENT_NAME.",
	RunE:  credentialsRequestProof,
	Args:  cobra.ExactArgs(1),
}

func init() {
	credentialsCmd.AddCommand(credentialsRequestProofCmd)
	credentialsRequestProofCmd.Flags().StringArrayVar(&attrValues, "attr", []string{}, "list of attributes needed in the presentation [KEY=WQL_QUERY] (can be repeated)")

	credentialsRequestProofCmd.Flags().StringVar(&subject, "subject", "", "external name of the holder of the credential")
	_ = credentialsRequestProofCmd.MarkFlagRequired("subject")

	credentialsRequestProofCmd.Flags().StringVar(&schemaName, "schema-name", "", "name of schema requested")
	_ = credentialsRequestProofCmd.MarkFlagRequired("schema-name")

	credentialsRequestProofCmd.Flags().StringVar(&comment, "comment", "", "optional comment for the credential offer")
}

func credentialsRequestProof(_ *cobra.Command, args []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	ctx := context.Background()

	//TODO:  fill out these fields from parameters
	agentName := args[0]
	var presentation = &common.RequestPresentation{
		Name: "Proof Name...",
	}

	req := &common.RequestPresentationRequest{
		AgentName:    agentName,
		ExternalId:   subject,
		Presentation: presentation,
	}

	_, err = cli.RequestPresentation(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "unable to request proof presentation for %s", agentName)
	}

	fmt.Printf("PRESENTATION REQUEST FOR %s MADE\n", agentName)
	return nil
}
