/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/client/canis"
	"github.com/scoir/canis/pkg/client/informer"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
)

type scheduler struct {
	client api.AdminClient
}

func main() {

	//err := indy.ResolveDID("PkygzecB8VwTf9jAMYKDrS")
	//log.Fatalln(err)
	didinfo := &identifiers.MyDIDInfo{
		DID:        "",
		Seed:       "",
		Cid:        true,
		MethodName: "ioe",
	}

	did, _, err := identifiers.CreateDID(didinfo)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("DID:", did.String())
	fmt.Println("Verkey:", did.AbbreviateVerkey())

	client := canis.New("127.0.0.1:7778")
	client.AgentInformer().AddEventHandler(informer.ResourceEventHandlerFuncs{
		AddFunc: func(n interface{}) {
			log.Println("*** ADDED ***")
			d, _ := json.MarshalIndent(n, " ", " ")
			fmt.Println(string(d))
		},
		DeleteFunc: func(n interface{}) {
			log.Println("*** DELETED ***")
			d, _ := json.MarshalIndent(n, " ", " ")
			fmt.Println(string(d))
		},
	})

	ch := make(chan bool)
	<-ch
}

func (r *scheduler) watchAgents() error {
	log.Println("trying agent watch")
	ctx, cancelFunc := context.WithCancel(context.Background())
	stream, err := r.client.WatchAgents(ctx, &api.WatchRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to connect")
	}

	go func() {
		time.Sleep(8 * time.Second)
		cancelFunc()
	}()

	for {
		evt, err := stream.Recv()
		fmt.Println(status.Code(err))
		if err == io.EOF || status.Code(err) == codes.Canceled {
			return nil
		}

		if err != nil {
			return err
		}

		d, _ := json.MarshalIndent(evt, " ", " ")
		fmt.Println(string(d))
	}

}
