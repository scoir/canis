/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	docutil "github.com/hyperledger/aries-framework-go/pkg/doc/util"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	mongodbstore "github.com/scoir/aries-storage-mongo/pkg/storage"
	goji "goji.io"
	"goji.io/pat"

	"github.com/scoir/canis/pkg/clr"
	"github.com/scoir/canis/pkg/credential"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/util"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var credsup *credential.Supervisor
var subject *didexchange.Connection
var credcl *issuecredential.Client
var offerID string

func main() {
	arieslog.SetLevel("aries-framework/out-of-band/service", arieslog.CRITICAL)
	arieslog.SetLevel("aries-framework/ws", arieslog.CRITICAL)
	//arieslog.SetLevel("aries-framework/did-exchange/service", arieslog.DEBUG)
	createAriesContext()
	listen()
}

func listen() {
	mux := goji.NewMux()
	mux.Handle(pat.Get("/invitation/:name"), http.HandlerFunc(invitation))
	mux.Handle(pat.Post("/issue-credential"), http.HandlerFunc(issueCredential))

	u := "0.0.0.0:2002"
	log.Println("issuer listening on", u)
	err := http.ListenAndServe(u, mux)
	log.Fatalln("issuer no longer listening", err)
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
	d, _ := json.MarshalIndent(conn, " ", " ")
	fmt.Println(string(d))
	log.Println("Connection to", id, "succeeded!")
}

func failed(id string, err error) {
	log.Println("Connection to", id, "failed with error:", err)
}

func createAriesContext() {
	wsinbound := "0.0.0.0:2001"

	ar, err := aries.New(
		aries.WithStoreProvider(mongodbstore.NewProvider("mongodb://172.17.0.1:27017", mongodbstore.WithDBPrefix("issuer"))),
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
	bouncer, _ = didex.NewBouncer(prov)
	credcl, _ = prov.GetCredentialClient()

	credsup, err = credential.New(prov)
	if err != nil {
		log.Fatalln("unable to create new credential supervisor", err)
	}

	handler := &credHandler{}
	err = credsup.Start(handler)
	if err != nil {
		log.Fatalln("unable to start credential supervisor", err)
	}
}

func issueCredential(w http.ResponseWriter, _ *http.Request) {

	if subject == nil {
		util.WriteError(w, "Invalid request, please connect with subject first")
		return
	}

	vc := generateCredential()

	offer := &issuecredential.OfferCredential{
		Comment: "High School Final Transcript",
		CredentialPreview: icprotocol.PreviewCredential{
			Type: "Clr",
			Attributes: []icprotocol.Attribute{
				{
					Name:  "Achievement",
					Value: "Mathmatics - Algebra Level 1",
				},
			},
		},
		OffersAttach: []decorator.Attachment{
			{Data: decorator.AttachmentData{JSON: vc}},
		},
	}

	id, err := credcl.SendOffer(offer, subject.MyDID, subject.TheirDID)
	if err != nil {
		util.WriteErrorf(w, "unable to offer credential to client: %v", err)
		return
	}

	offerID = id

	d, err := json.Marshal(struct{ OfferId string }{id})
	if err != nil {
		util.WriteErrorf(w, "unable to marshal offer: %v", err)
		return
	}

	util.WriteSuccess(w, d)
	return
}

func generateCredential() *verifiable.Credential {
	var issued = time.Date(2010, time.January, 1, 19, 23, 24, 0, time.UTC)

	record := &clr.CLR{
		Context: []string{
			"https://purl.imsglobal.org/spec/clr/v1p0/context/clr_v1p0.jsonld",
		},
		ID:   "did:scoir:abc123",
		Type: "Clr",
		Learner: &clr.Profile{
			ID:    "did:scoir:hss123",
			Type:  "Profile",
			Email: "student1@highschool.k12.edu",
		},
		Publisher: &clr.Profile{
			ID:    "did:scoir:highschool",
			Type:  "Profile",
			Email: "counselor@highschool.k12.edu",
		},
		Assertions: []*clr.Assertion{
			{
				ID:   "did:scoir:assert123",
				Type: "Assertion",
				Achievement: &clr.Achievement{
					ID:              "did:scoir:achieve123",
					AchievementType: "Achievement",
					Name:            "Mathmatics - Algebra Level 1",
				},
				IssuedOn: docutil.NewTime(issued),
			},
		},
		Achievements: nil,
		IssuedOn:     docutil.NewTime(issued),
	}

	vc := &verifiable.Credential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://purl.imsglobal.org/spec/clr/v1p0/context/clr_v1p0.jsonld",
		},
		ID: "http://example.edu/credentials/1872",
		Types: []string{
			"VerifiableCredential",
			"Clr"},
		Subject: record,
		Issuer: verifiable.Issuer{
			ID: subject.MyDID,
		},
		Issued:  docutil.NewTime(issued),
		Schemas: []verifiable.TypedID{},
		CustomFields: map[string]interface{}{
			"referenceNumber": 83294847,
		},
	}

	signCred(vc)
	return vc
}

func signCred(vc *verifiable.Credential) {

	doc, err := ctx.VDRIRegistry().Resolve(subject.MyDID)
	if err != nil {
		log.Fatalln("unable to load my did doc")
	}

	signer, err := newCryptoSigner(doc.PublicKey[0].ID[1:])
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("ID", doc.PublicKey[0].ID)
	sigSuite := ed25519signature2018.New(
		suite.WithSigner(signer),
		suite.WithVerifier(ed25519signature2018.NewPublicKeyVerifier()))

	ldpContext := &verifiable.LinkedDataProofContext{
		SignatureType:           "Ed25519Signature2018",
		SignatureRepresentation: verifiable.SignatureProofValue,
		Suite:                   sigSuite,
		VerificationMethod:      fmt.Sprintf("%s%s", subject.MyDID, doc.PublicKey[0].ID),
	}

	err = vc.AddLinkedDataProof(ldpContext)
	if err != nil {
		log.Fatalln(err)
	}

}

func newCryptoSigner(kid string) (*subtle.ED25519Signer, error) {
	priv, err := ctx.KMS().Get(kid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find key set")
	}

	kh := priv.(*keyset.Handle)
	prim, err := kh.Primitives()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load signer primitives")
	}
	return prim.Primary.Primitive.(*subtle.ED25519Signer), nil

}

type credHandler struct{}
type prop interface {
	MyDID() string
	TheirDID() string
}

func (c *credHandler) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {
	panic("implement me")
}

func (c *credHandler) OfferCredentialMsg(_ service.DIDCommAction, _ *icprotocol.OfferCredential) {
	panic("implement me")
}

func (c *credHandler) IssueCredentialMsg(_ service.DIDCommAction, _ *icprotocol.IssueCredential) {
	panic("implement me")
}

func (c *credHandler) RequestCredentialMsg(e service.DIDCommAction, request *icprotocol.RequestCredential) {
	props := e.Properties.(prop)
	theirDID := props.TheirDID()

	if theirDID != subject.TheirDID {
		log.Println("invalid request for credential by", theirDID)
		return
	}

	for _, attach := range request.RequestsAttach {
		cred, _ := attach.Data.JSON.(map[string]interface{})
		id, _ := cred["id"].(string)
		if id == "" {
			log.Println("no ID found in request attachment")
			continue
		}

		var msg *icprotocol.IssueCredential
		thid, _ := e.Message.ThreadID()

		fmt.Printf("offerID: %s, credID: %s, threadID: %s\n", offerID, id, thid)

		if offerID == thid {
			msg = &icprotocol.IssueCredential{
				Type:    icprotocol.IssueCredentialMsgType,
				Comment: fmt.Sprintf("CLR Transcript"),
				CredentialsAttach: []decorator.Attachment{
					{Data: decorator.AttachmentData{JSON: generateCredential()}},
				},
			}

			//TODO:  Shouldn't this be built into the Supervisor??
			log.Println("setting up monitoring for", thid)
			mon := credential.NewMonitor(credsup)
			mon.WatchThread(thid, TranscriptAccepted(id), CredentialError)
		}
		e.Continue(icprotocol.WithIssueCredential(msg))

	}
}

func TranscriptAccepted(id string) func(threadID string, ack *model.Ack) {

	return func(threadID string, ack *model.Ack) {
		fmt.Printf("Transcript Accepted: %s", id)
	}
}

func CredentialError(threadID string, err error) {
	log.Println("step 1... failed!", threadID, err)
}
