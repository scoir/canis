/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"
	"strings"

	"github.com/google/tink/go/keyset"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"

	"github.com/scoir/canis/pkg/aries/storage/mongodb/store"
	"github.com/scoir/canis/pkg/aries/vdri/indy"
)

func main() {
	indyVDRI, err := indy.New("scoir", indy.WithIndyVDRGenesisFile("/home/pfeairheller/git_root/canis/genesis.txn"))
	if err != nil {
		log.Fatalln("new indy", err)
	}

	mlk := "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ="
	lock, err := local.NewService(strings.NewReader(mlk), nil)
	if err != nil {
		log.Fatalln(err)
	}

	mongodb := store.NewProvider("mongodb://172.17.0.1:27017", "canis")
	framework, err := aries.New(
		aries.WithStoreProvider(mongodb),
		aries.WithVDRI(indyVDRI),
		aries.WithSecretLock(lock),
	)
	if err != nil {
		log.Fatalln("new framework", err)
	}

	ctx, err := framework.Context()
	if err != nil {
		log.Fatalln("get context", err)
	}

	priv, err := ctx.KMS().Get("mSrVQRv0MJL7MWAWbJUYs2tR7THnsKBybjyW_sBJQXY")
	if err != nil {
		log.Fatalln("no key", err)
	}

	log.Println((priv.(*keyset.Handle)).String())

}
