/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/hyperledger/aries-framework-go/pkg/vdri/indy"
)

func main() {
	indyVDRI, err := indy.New("scoir", indy.WithIndyVDRGenesisFile("/home/pfeairheller/git_root/canis/genesis.txn"))
	if err != nil {
		log.Fatalln("new indy", err)
	}
	framework, err := aries.New(
		aries.WithStoreProvider(mem.NewProvider()),
		aries.WithVDRI(indyVDRI),
	)
	if err != nil {
		log.Fatalln("new framework", err)
	}

	ctx, err := framework.Context()
	if err != nil {
		log.Fatalln("get context", err)
	}

	registry := ctx.VDRIRegistry()

	doc, err := registry.Create("scoir", vdri.WithServiceEndpoint("http://69.69.69.69:6969"))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(doc.ID)
}
