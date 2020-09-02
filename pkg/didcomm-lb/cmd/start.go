/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	couchdbstore "github.com/hyperledger/aries-framework-go/pkg/storage/couchdb"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm load balancer",
	Long:  `Starts a didcomm load balancer`,
	Run:   runStart,
}

func runStart(cmd *cobra.Command, args []string) {
	addr := fmt.Sprintf("%s:%s", ctx.vp.GetString("inbound.host"), ctx.vp.GetString("inbound.port"))

	srv := &http.Server{Addr: addr}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if valid := validateHTTPMethod(w, r); !valid {
			return
		}

		if valid := validatePayload(r, w); !valid {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read payload", http.StatusInternalServerError)
			return
		}

		cdb, err := couchdbstore.NewProvider("localhost:5984")
		if err != nil {
			log.Println("couchdb provider", err)
			return
		}

		framework, err := aries.New(aries.WithStoreProvider(cdb))
		if err != nil {
			log.Println("aries", err)
			return
		}

		ctx, err := framework.Context()
		if err != nil {
			log.Println("framework context", err)
			return
		}

		_, err = ctx.Packager().UnpackMessage(body)
		if err != nil {
			//http.Error(w, "failed to unpack msg", http.StatusInternalServerError)
			//return
		}
		//ignore for now
		//b, err := json.Marshal(unpackMsg)
		//fmt.Println(string(b))

		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			log.Println("amqp dial", err)
			return
		}

		defer func() {
			err := conn.Close()
			if err != nil {
				log.Println("failed to close mq", err)
			}
		}()

		ch, err := conn.Channel()
		if err != nil {
			log.Println("channel", err)
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
			log.Println("queue declare", err)
			return
		}
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				//This should be b ignored from above
				Body: body,
			})
		if err != nil {
			return
		}
	})

	srv.Handler = handler

	_ = srv.ListenAndServe()
}

func validatePayload(r *http.Request, w http.ResponseWriter) bool {
	if r.ContentLength == 0 { // empty payload should not be accepted
		http.Error(w, "Empty payload", http.StatusBadRequest)
		return false
	}

	return true
}

func validateHTTPMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	ct := r.Header.Get("Content-type")
	if ct != "application/didcomm-envelope-enc" {
		log.Println("hihihihi")
		http.Error(w, fmt.Sprintf("Unsupported Content-type \"%s\"", ct), http.StatusUnsupportedMediaType)
		return false
	}

	return true
}

func init() {
	rootCmd.AddCommand(startCmd)
}
