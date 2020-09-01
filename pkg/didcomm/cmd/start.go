/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm service",
	Long:  `Starts a didcomm service`,
	Run:   runStart,
}

func runStart(cmd *cobra.Command, args []string) {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer func() {
		err := ch.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	q, err := ch.QueueDeclare(
		"lb-q", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			//make aries go
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func init() {
	rootCmd.AddCommand(startCmd)
}
