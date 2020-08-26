/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/duration"
)

var agents bool

// startCmd represents the start command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of a running canis credential hub",
	Long:  `Gets the status of a running canis credential hub`,
	Run:   displayStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&agents, "agents", false, "include status for agent processes")
}

func displayStatus(cmd *cobra.Command, _ []string) {
	executor, err := ctx.Executor()
	if err != nil {
		log.Fatalln("unable to access executor", err)
	}

	//Check Core Component Status
	processes := executor.PS()

	tab := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	cmd.SetOut(tab)

	cmd.Println("NAME\tID\tSTATUS\tTIME")

	for _, p := range processes {
		cmd.Print(p.Name(), "\t", p.ID(), "\t", p.Status(), "\t", duration.HumanDuration(p.Time()), "\n")
	}
	_ = tab.Flush()

	aps := executor.AgentPS()
	if agents {
		for _, p := range aps {
			cmd.Print(p.ID(), "\t", p.Status(), "\t", duration.HumanDuration(p.Time()), "\n")
		}
		_ = tab.Flush()
	} else {
		cmd.Printf("\nAgents running: %d\n", len(aps))
	}

}
