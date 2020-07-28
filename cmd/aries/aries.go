package main

import (
	"encoding/json"
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

	//doc, err := registry.Resolve("did:scoir:Xy9dvEi8dkkPif5j342w9q")
	//if err != nil {
	//	log.Fatalln("resolve did", err)
	//}
	//
	//d, _ := json.MarshalIndent(doc, " ", " ")
	//fmt.Println(string(d))
	//
	doc, err := registry.Create("scoir", vdri.WithServiceEndpoint("http://69.69.69.69:6969"))
	if err != nil {
		log.Fatalln(err)
	}

	d, _ := json.MarshalIndent(doc, " ", " ")
	fmt.Println(string(d))
}
