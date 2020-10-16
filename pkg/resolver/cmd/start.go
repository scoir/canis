/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/resolver"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the http-indy-resolver",
	Long:  `Starts the http-indy-resolver`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {

	log.Println("starting http-indy-resolver")

	endpoint, err := prov.GetHTTPEndpoint()
	if err != nil {
		log.Fatalln("endpoint not configured correctly", err)
	}

	method := prov.vp.GetString("resolver.method")

	srv := resolver.NewHTTPIndyResolver(endpoint.Address(), method, prov)
	err = srv.Start()
	if err != nil {
		log.Fatalln("unable to launch http-indy-resolver", err)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
