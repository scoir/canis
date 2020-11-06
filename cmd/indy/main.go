package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/indy-vdr/wrappers/golang/crypto"
	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/mr-tron/base58"
)

func main() {
	genesis, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("unable to open genesis file", err)
	}

	client, err := vdr.New(genesis)
	if err != nil {
		log.Fatalln(err)
	}

	err = client.RefreshPool()
	if err != nil {
		log.Fatalln(err)
	}

	//status, err := client.GetPoolStatus()
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ := json.MarshalIndent(status, " ", " ")
	//fmt.Println(string(d))
	//
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

	var pubkey ed25519.PublicKey
	var privkey ed25519.PrivateKey
	privkey = ed25519.NewKeyFromSeed([]byte("b2352b32947e188eb72871093ac6217e"))
	pubkey = privkey.Public().(ed25519.PublicKey)
	writerDID, err := identifiers.CreateDID(&identifiers.MyDIDInfo{PublicKey: pubkey, Cid: true, MethodName: "sov"})
	if err != nil {
		log.Fatalln(err)
	}

	mysig := crypto.NewSigner(pubkey, privkey)
	someRandomPubkey, someRandomPrivkey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalln(err)
	}

	someRandomDID, err := identifiers.CreateDID(&identifiers.MyDIDInfo{PublicKey: someRandomPubkey, MethodName: "sov", Cid: true})
	if err != nil {
		log.Fatalln(err)
	}

	rply, err := client.GetTxnAuthorAgreement()
	if err != nil {
		log.Fatalln(err)
	}

	taa := vdr.TransacriptAuthorAgreement{}
	err = taa.Unmarshal(rply)

	rply, err = client.GetAcceptanceMethodList()
	if err != nil {
		log.Fatalln(err)
	}

	aml := &vdr.AcceptanceMethodList{}
	err = aml.Unmarshal(rply)
	if err != nil {
		log.Fatalln(err)
	}

	acceptance := "at_submission"

	err = client.CreateNym(
		someRandomDID.DIDVal.MethodSpecificID,
		someRandomDID.Verkey, vdr.NoRole,
		writerDID.DIDVal.MethodSpecificID, mysig, vdr.WithTAA(acceptance, taa.Digest, taa.RatificationTS))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("New Endorser DID:", someRandomDID.String())
	fmt.Println("New Endorser Verkey:", someRandomDID.AbbreviateVerkey())
	fmt.Println("Place These in Wallet:")
	fmt.Println("Public:", base58.Encode(someRandomPubkey))
	fmt.Println("Private:", base58.Encode(someRandomPrivkey))

	//rply, err = client.GetCredDef("Xy9dvEi8dkkPif5j342w9q:3:CL:23:default")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//d, _ = json.MarshalIndent(rply, " ", " ")
	//fmt.Println(string(d))

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
