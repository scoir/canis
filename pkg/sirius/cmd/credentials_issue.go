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

var credentialsIssueCmd = &cobra.Command{
	Use:   "issue AGENT_NAME",
	Short: "Issue a schema on behalf of AGENT_NAME.",
	RunE:  credentialsIssue,
	Args:  cobra.ExactArgs(1),
}

var schemaName string
var comment string

func init() {
	credentialsCmd.AddCommand(credentialsIssueCmd)
	credentialsIssueCmd.Flags().StringArrayVar(&attrValues, "attr", []string{}, "list of attributes for this schema [KEY=VALUE] (can be repeated)")

	credentialsIssueCmd.Flags().StringVar(&subject, "subject", "", "external name of the holder of the credential")
	_ = credentialsIssueCmd.MarkFlagRequired("subject")

	credentialsIssueCmd.Flags().StringVar(&schemaName, "schema-name", "", "name of schema to issue")
	_ = credentialsIssueCmd.MarkFlagRequired("schema-id")

	credentialsIssueCmd.Flags().StringVar(&comment, "comment", "", "optional comment for the credential offer")
}

func credentialsIssue(_ *cobra.Command, args []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	ctx := context.Background()

	preview := make([]*common.CredentialAttribute, len(attrValues))
	for i, value := range attrValues {
		vals := strings.Split(value, "=")
		if len(vals) != 2 {
			return errors.Errorf("invalid attribute %s, must be in format [name=value]", value)
		}

		preview[i] = &common.CredentialAttribute{
			Name:  vals[0],
			Value: vals[1],
		}
	}

	agentName := args[0]
	credential := &common.Credential{
		SchemaId: schemaName,
		Comment:  comment,
		Type:     schemaType,
		Preview:  preview,
	}
	req := &common.IssueCredentialRequest{
		AgentName:  agentName,
		ExternalId: subject,
		Credential: credential,
	}

	_, err = cli.IssueCredential(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "unable to issue credential for %s", agentName)
	}

	fmt.Printf("CREDENTIAL ISSUED FOR %s\n", agentName)
	return nil
}
