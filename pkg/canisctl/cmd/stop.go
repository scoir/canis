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
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// stopCmd represents the start command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Starts a canis credential hub",
	Long:  `Starts a canis credential hub.`,
	Run:   stopCluster,
}

func stopCluster(_ *cobra.Command, _ []string) {
	executor, err := frameworkCfg.Execution.Executor()
	if err != nil {
		log.Fatalln("unable to access executor", err)
	}

	err = executor.ShutdownSteward()
	if err != nil {
		log.Println("error stopping steward", err)
	}

	fmt.Println("cluster is stopped")
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
