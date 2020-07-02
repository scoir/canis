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
	"gopkg.in/yaml.v2"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a canis credential hub",
	Long:  `Starts a canis credential hub.`,
	Run:   startCluster,
}

func startCluster(_ *cobra.Command, _ []string) {
	executor, err := frameworkCfg.Execution.Executor()
	if err != nil {
		log.Fatalln("unable to access executor", err)
	}

	ex := config.GetStringMap("execution")
	ds := config.GetStringMap("datastore")
	st := config.Sub("steward")

	st.Set("execution", ex)
	st.Set("datastore", ds)

	out := map[string]interface{}{}
	err = st.Unmarshal(&out)

	d, err := yaml.Marshal(out)
	if err != nil {
		log.Fatalln("unable to marshal steward config")
	}

	pid, err := executor.LaunchSteward(d)
	if err != nil {
		log.Println("error launching steward", err)
	}

	fmt.Println("steward launched at", pid)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
