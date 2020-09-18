/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/controller"
	lb "github.com/scoir/canis/pkg/didcomm/loadbalancer"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm load balancer",
	Long:  `Starts a didcomm load balancer`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	host := prov.vp.GetString("inbound.host")
	httpPort := prov.vp.GetInt("inbound.httpport")
	wsPort := prov.vp.GetInt("inbound.wsport")
	external := prov.vp.GetString("inbound.external")

	log.Println("starting didcomm loadbalancer")

	ar, err := aries.New(
		aries.WithStoreProvider(prov.ariesStorageProvider),
		aries.WithSecretLock(prov.lock),
	)
	if err != nil {
		log.Fatalln("unable to initialize aries", err)
	}

	ctx, err := ar.Context()
	if err != nil {
		log.Fatalln("unable to get aries context", err)
	}

	srv, err := lb.New(ctx, prov.GetAMQPAddress(), host, httpPort, wsPort, external)
	if err != nil {
		log.Fatalln("unable to launch didcomm loadbalancer ")
	}

	srv.Start()

	runner, err := controller.New(prov, srv)
	if err != nil {
		log.Fatalln("unable to start didcomm-loadbalancer", err)
	}

	err = runner.Launch()
	if err != nil {
		log.Fatalln("launch errored with", err)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
