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

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
	"github.com/scoir/canis/pkg/util"
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

	cc, err := grpc.Dial("127.0.0.1:7778", grpc.WithInsecure())
	if err != nil {
		log.Fatalln("can't connect", err)
	}
	s := &scheduler{
		client: api.NewAdminClient(cc),
	}

	err = backoff.RetryNotify(s.watchAgents, backoff.NewExponentialBackOff(), util.Logger)
}

func (r *scheduler) watchAgents() error {
	log.Println("trying agent watch")
	stream, err := r.client.WatchAgents(context.Background(), &api.WatchRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to connect")
	}

	for {
		evt, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		d, _ := json.MarshalIndent(evt, " ", " ")
		fmt.Println(string(d))
	}

}
