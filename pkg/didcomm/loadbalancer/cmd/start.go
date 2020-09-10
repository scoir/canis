/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/spf13/cobra"

	lb "github.com/scoir/canis/pkg/didcomm/loadbalancer"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm load balancer",
	Long:  `Starts a didcomm load balancer`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	amqpUser := prov.vp.GetString("amqp.user")
	amqpPwd := prov.vp.GetString("amqp.password")
	amqpHost := prov.vp.GetString("amqp.host")
	amqpPort := prov.vp.GetInt("amqp.port")
	amqpVHost := prov.vp.GetString("amqp.vhost")
	amqpAddr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", amqpUser, amqpPwd, amqpHost, amqpPort, amqpVHost)
	host := prov.vp.GetString("inbound.host")
	httpPort := prov.vp.GetInt("inbound.httpport")
	wsPort := prov.vp.GetInt("inbound.wsport")

	wait := make(chan bool)

	log.Println("starting didcomm loadbalancer")

	ar, err := aries.New(
		aries.WithStoreProvider(prov.ariesStorageProvider),
	)
	if err != nil {
		log.Fatalln("unable to initialize aries", err)
	}

	ctx, err := ar.Context()
	if err != nil {
		log.Fatalln("unable to get aries context", err)
	}

	srv, err := lb.New(ctx, amqpAddr, host, httpPort, wsPort)
	if err != nil {
		log.Fatalln("unable to launch didcomm loadbalancer ")
	}

	srv.Start()

	<-wait
}

func init() {
	rootCmd.AddCommand(startCmd)
}
