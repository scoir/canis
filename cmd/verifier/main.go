package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	goji "goji.io"
	"goji.io/pat"

	"github.com/scoir/canis/pkg/aries/storage/mongodb/store"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/util"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var subject *didexchange.Connection

func main() {
	createAriesContext()
	listen()
}

func listen() {
	mux := goji.NewMux()
	mux.Handle(pat.Get("/invitation/:name"), http.HandlerFunc(invitation))
	mux.Handle(pat.Post("/request-proof"), http.HandlerFunc(requestProof))

	u := "0.0.0.0:4002"
	log.Println("verifier listening on", u)
	err := http.ListenAndServe(u, mux)
	log.Fatalln("verifier no longer listening", err)
}

func requestProof(w http.ResponseWriter, req *http.Request) {
}

func invitation(w http.ResponseWriter, req *http.Request) {
	name := pat.Param(req, "name")
	invite, err := bouncer.CreateInvitationNotify(name, accepted, failed)
	if err != nil {
		util.WriteError(w, err.Error())
		return
	}

	d, _ := json.MarshalIndent(invite, " ", " ")
	fmt.Println(string(d))
	util.WriteSuccess(w, d)
}

func accepted(id string, conn *didexchange.Connection) {
	subject = conn
	log.Println("Connection to", id, "succeeded!")
}

func failed(id string, err error) {
	log.Println("Connection to", id, "failed with error:", err)
}

func createAriesContext() {
	wsinbound := "0.0.0.0:4001"

	ar, err := aries.New(
		aries.WithStoreProvider(store.NewProvider("mongodb://172.17.0.1:27017", "verifier")),
		defaults.WithInboundWSAddr(wsinbound, wsinbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
	)
	if err != nil {
		log.Fatalln("Unable to create", err)
	}

	ctx, err = ar.Context()
	if err != nil {
		log.Fatalln("unable to get context", err)
	}

	prov := framework.NewSimpleProvider(ctx)
	bouncer, err = didex.NewBouncer(prov)
}
