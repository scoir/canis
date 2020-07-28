/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"

	did2 "github.com/scoir/canis/pkg/did"
)

func main() {

	//err := indy.ResolveDID("PkygzecB8VwTf9jAMYKDrS")
	//log.Fatalln(err)
	didinfo := &did2.MyDIDInfo{
		DID:        "",
		Seed:       "",
		Cid:        true,
		MethodName: "ioe",
	}

	did, keypair, err := did2.CreateMyDid(didinfo)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("DID:", did)
	fmt.Println("Verkey:", keypair.Verkey())
}
