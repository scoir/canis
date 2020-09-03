/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"

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
	amqpUser := ctx.vp.GetString("amqp.user")
	amqpPwd := ctx.vp.GetString("amqp.password")
	amqpHost := ctx.vp.GetString("amqp.host")
	amqpPort := ctx.vp.GetInt("amqp.port")
	amqpVHost := ctx.vp.GetString("amqp.vhost")
	amqpAddr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", amqpUser, amqpPwd, amqpHost, amqpPort, amqpVHost)
	host := ctx.vp.GetString("inbound.host")
	httpPort := ctx.vp.GetInt("inbound.httpport")
	wsPort := ctx.vp.GetInt("inbound.wsport")

	wait := make(chan bool)

	log.Println("starting didcomm loadbalancer")
	srv, err := lb.New(amqpAddr, host, httpPort, wsPort)
	if err != nil {
		log.Fatalln("unable to launch didcomm loadbalancer ")
	}

	srv.Start()

	<-wait
}

func init() {
	rootCmd.AddCommand(startCmd)
}
