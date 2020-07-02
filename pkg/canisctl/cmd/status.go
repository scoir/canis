/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
	executor, err := frameworkCfg.Execution.Executor()
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
