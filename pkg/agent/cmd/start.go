/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/agent"
	"github.com/scoir/canis/pkg/controller"
)

var agentID string
var publicDID bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the agent",
	Long:  `Starts the agent`,
	Run:   runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&agentID, "id", "", "unique, installation specific ID for this agent")
	startCmd.Flags().BoolVar(&publicDID, "public", false, "should this agent register a public DID")
	_ = startCmd.MarkFlagRequired("id")
	_ = startCmd.MarkFlagRequired("public")
}

func runStart(_ *cobra.Command, _ []string) {

	a, err := agent.NewAgent(ctx, agent.WithAgentID(agentID), agent.WithPublicDID(publicDID))
	if err != nil {
		log.Fatalln("error initializing agent", err)
	}

	runner, err := controller.New(ctx, a)
	if err != nil {
		log.Fatalln("unable to start steward", err)
	}

	err = runner.Launch()
	if err != nil {
		log.Fatalln("launch errored with", err)
	}

	log.Println("Shutdown")
}
