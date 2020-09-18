package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	goji "goji.io"
	"goji.io/pat"

	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	canisproof "github.com/scoir/canis/pkg/presentproof"
	"github.com/scoir/canis/pkg/util"
	mongodbstore "github.com/scoir/aries-storage-mongo/pkg/storage"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var subject *didexchange.Connection
var proof *ppclient.Client
var pofHandler *proofHandler
var proofID string

func main() {
	arieslog.SetLevel("aries-framework/out-of-band/service", arieslog.CRITICAL)
	arieslog.SetLevel("aries-framework/ws", arieslog.CRITICAL)
	//arieslog.SetLevel("aries-framework/did-exchange/service", arieslog.DEBUG)
	//arieslog.SetLevel("aries-framework/issuecredential/service", arieslog.DEBUG)
	arieslog.SetLevel("aries-framework/presentproof/service", arieslog.DEBUG)

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

func requestProof(w http.ResponseWriter, _ *http.Request) {
	if subject == nil {
		util.WriteError(w, "please connect to subject before requesting proof")
		return
	}

	var err error
	presentationReq := &ppclient.RequestPresentation{
		Comment:                    "",
		Formats:                    nil,
		RequestPresentationsAttach: nil,
	}
	proofID, err = proof.SendRequestPresentation(presentationReq, subject.MyDID, subject.TheirDID)
	if err != nil {
		util.WriteErrorf(w, "unable to send presentation request to subject: %v", err)
		return
	}

	d, err := json.Marshal(struct{ ProofID string }{proofID})
	if err != nil {
		util.WriteErrorf(w, "unable to marshal offer: %v", err)
		return
	}

	util.WriteSuccess(w, d)
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
		aries.WithStoreProvider(mongodbstore.NewProvider("mongodb://172.17.0.1:27017", mongodbstore.WithDBPrefix("verifier"))),
		defaults.WithInboundWSAddr(wsinbound, fmt.Sprintf("ws://%s", wsinbound), "", ""),
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
	if err != nil {
		log.Fatalln("unable to initialize bounder", err)
	}

	proof, err = ppclient.New(ctx)
	if err != nil {
		log.Fatalln("unable to initialize proof client", err)
	}

	pofHandler, err = NewProofHandler(ctx)
	if err != nil {
		log.Fatalln("unable to create proof handler", err)
	}

	psup, err := canisproof.New(prov)
	if err != nil {
		log.Fatalln("unable to create proof supervisor", err)
	}

	err = psup.Start(pofHandler)
	if err != nil {
		log.Fatalln("unable to start proof supervisor", err)
	}

}

func NewProofHandler(ctx *ariescontext.Provider) (*proofHandler, error) {

	ppcl, err := ppclient.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	a := &proofHandler{
		ppcl: ppcl,
	}

	return a, nil
}

type prop interface {
	PIID() string
}

type proofHandler struct {
	ppcl *ppclient.Client
}

func (r *proofHandler) ProposePresentationMsg(_ service.DIDCommAction, _ *presentproof.ProposePresentation) {
}

func (r *proofHandler) RequestPresentationMsg(_ service.DIDCommAction, _ *presentproof.RequestPresentation) {
}

func (r *proofHandler) PresentationMsg(e service.DIDCommAction, pres *presentproof.Presentation) {
	fmt.Printf("Accepting presentation %s:\n", pres.PresentationsAttach[0].ID)
	p := e.Properties.(prop)
	err := r.ppcl.AcceptPresentation(p.PIID())
	if err != nil {
		log.Println("err accepting presentation", err)
	}

	log.Println("succeeded accepting presentation")
}

func (r *proofHandler) PresentationPreviewMsg(e service.DIDCommAction, d *presentproof.Presentation) {
	panic("implement me")
}
