package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"

	"github.com/scoir/canis/pkg/schema"
)

func main() {

	nonce, err := ursa.NewNonce()
	if err != nil {
		log.Fatalln("error", err)
	}

	js, _ := nonce.ToJSON()
	fmt.Println(string(js))
	fmt.Println(strings.Trim(string(js), "\""))

	_, err = ursa.NonceFromJSON(string(js))
	if err != nil {
		log.Println("handle error", err)
	}

	genesis, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("unable to open genesis file", err)
	}

	client, err := vdr.New(genesis)
	if err != nil {
		log.Fatalln(err)
	}
	//
	//err = client.RefreshPool()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//status, err := client.GetPoolStatus()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ := json.MarshalIndent(status, " ", " ")
	//fmt.Println(string(d))

	//rply, err := client.GetNym("Xy9dvEi8dkkPif5j342w9q")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//m := map[string]interface{}{}
	//err = json.Unmarshal([]byte(rply.Data.(string)), &m)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//fmt.Println(m["dest"])
	//fmt.Println(m["verkey"])

	//rply, err := client.GetTxnAuthorAgreement()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ = json.MarshalIndent(rply, " ", " ")
	//fmt.Println(string(d))
	//
	//rply, err = client.GetAcceptanceMethodList()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ = json.MarshalIndent(rply, " ", " ")
	//fmt.Println(string(d))
	//
	rply, err := client.GetCredDef("Xy9dvEi8dkkPif5j342w9q:3:CL:23:default")
	if err != nil {
		log.Fatalln(err)
	}

	indyCredDef := &schema.IndyCredDef{}
	d, _ := json.Marshal(rply.Data)

	err = json.Unmarshal(d, indyCredDef)
	if err != nil {
		log.Fatalln(err)
	}

	d, _ = json.MarshalIndent(indyCredDef, " ", " ")
	fmt.Println(string(d))

	//rply, err = client.GetSchema("Xy9dvEi8dkkPif5j342w9q:2:Scoir High School:0.0.1")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ = json.MarshalIndent(rply, " ", " ")
	//fmt.Println(string(d))
	//
	//rply, err = client.GetEndpoint("Xy9dvEi8dkkPif5j342w9q")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ = json.MarshalIndent(rply, " ", " ")
	//fmt.Println(string(d))
	//
}
