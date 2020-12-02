/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	api "github.com/scoir/canis/pkg/apiserver/api/protogen"
)

var seed string

// serverInitCmd represents the init command
var serverInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the canis instance public key",
	Long:  `Initialize the canis instance public key`,
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

	_, err = cli.SeedPublicDID(context.Background(), &api.SeedPublicDIDRequest{Seed: seed})
	if err != nil {
		log.Fatalln("error seeding public DID for API server", err)
	}

	log.Println("apiserver seed successful")
}
