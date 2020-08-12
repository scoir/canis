/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/apiserver/api"
)

var seed string

// serverInitCmd represents the init command
var serverInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the steward's public key in a canis credential hub",
	Long:  `Initialize the steward's public key in a canis credential hub`,
	Run:   initCluster,
}

func init() {
	rootCmd.AddCommand(serverInitCmd)
	serverInitCmd.Flags().StringVar(&seed, "seed", "", "seed for public DID")
}

func initCluster(_ *cobra.Command, _ []string) {
	if seed == "" {
		log.Fatalln("seed flag is required")
	}

	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	_, err = cli.SeedPublicDID(context.Background(), &api.SeePublicDIDRequest{Seed: seed})
	if err != nil {
		log.Fatalln("error seeding public DID for API server", err)
	}

	log.Println("apiserver seed successful")
}
