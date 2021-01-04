/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/protogen/common"
)

var file string
var url string

var invitationAcceptCmd = &cobra.Command{
	Use:       "accept AGENT_NAME",
	Short:     "Accept a connection invitation from another agent.",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"AGENT_NAME"},
	RunE:      invitationsAccept,
}

func init() {
	invitationsCmd.AddCommand(invitationAcceptCmd)

	invitationAcceptCmd.Flags().StringVar(&subject, "subject", "", "subject name for the destination of the invitation")
	_ = invitationAcceptCmd.MarkFlagRequired("subject")

	invitationAcceptCmd.Flags().StringVarP(&file, "file", "f", "", "name of file that contains a valid JSON invitation")

	invitationAcceptCmd.Flags().StringVarP(&url, "url", "u", "", "name of file that contains a valid JSON invitation")
}

func invitationsAccept(_ *cobra.Command, args []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	invitation, err := readInvitation()
	if err != nil {
		return errors.Wrap(err, "error getting invitation")
	}

	ctx := context.Background()
	req := &common.AcceptInvitationRequest{
		AgentName:  args[0],
		ExternalId: subject,
		Name:       "",
		Invitation: string(invitation),
	}

	_, err = cli.AcceptInvitation(ctx, req)
	if err != nil {
		return errors.Wrap(err, "unable to accept invitation")
	}

	return nil
}

func readInvitation() ([]byte, error) {
	if file != "" {
		reader, err := os.Open(file)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read invitation file")
		}
		defer reader.Close()
		return ioutil.ReadAll(reader)
	}

	if url != "" {
		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read invitation URL")
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("error %d reading invitation URL", resp.StatusCode)
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}

	return ioutil.ReadAll(os.Stdin)
}
