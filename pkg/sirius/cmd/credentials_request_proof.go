/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	attrs := make(map[string]*common.AttrInfo)
	for _, value := range attrValues {
		vals := strings.Split(value, "=")
		if len(vals) != 2 {
			return errors.Errorf("invalid attribute %s, must be in format [name=value]", value)
		}
		name := vals[0]

		attrs[name] = &common.AttrInfo{
			Name:         name,
			Restrictions: vals[1],
			NonRevoked:   nil,
		}
	}

	agentName := args[0]
	var presentation = &common.RequestPresentation{
		Name:                "Proof Name...",
		Version:             "0.1.0",
		SchemaId:            schemaName,
		Comment:             comment,
		WillConfirm:         true,
		RequestedAttributes: attrs,
		RequestedPredicates: nil,
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
