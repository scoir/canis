/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/dispatcher"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	vstore "github.com/hyperledger/aries-framework-go/pkg/store/verifiable"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"
	mongodbstore "github.com/scoir/aries-storage-mongo/pkg"
	goji "goji.io"
	"goji.io/pat"

	"github.com/scoir/canis/pkg/aries/vdri/indy"
	"github.com/scoir/canis/pkg/credential"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	canisproof "github.com/scoir/canis/pkg/presentproof"
	"github.com/scoir/canis/pkg/ursa"
	"github.com/scoir/canis/pkg/util"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var issuerConnection *didexchange.Connection
var verifierConnection *didexchange.Connection
var credHandler *credentialHandler
var pofHandler *proofHandler
var prover *ursa.Prover
var vdrclient *vdr.Client
var subjectStore storage.Store

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
	mux.Handle(pat.Post("/connect-to-issuer"), http.HandlerFunc(connectToIssuer))
	mux.Handle(pat.Post("/connect-to-verifier"), http.HandlerFunc(connectToVerifier))
	mux.Handle(pat.Get("/credentials"), http.HandlerFunc(getCredentials))

	u := "0.0.0.0:3002"
	log.Println("subject listening on", u)
	err := http.ListenAndServe(u, mux)
	log.Fatalln("subject no longer listening", err)
}

func getCredentials(w http.ResponseWriter, _ *http.Request) {
	creds, err := credHandler.GetCredentials()
	if err != nil {
		util.WriteErrorf(w, "unable to get credentials from store: %v", err)
		return
	}

	d, _ := json.MarshalIndent(creds, " ", " ")
	util.WriteSuccess(w, d)
}

func connectToIssuer(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	invite := &didexchange.Invitation{}
	err := json.NewDecoder(req.Body).Decode(invite)
	if err != nil {
		util.WriteErrorf(w, "Error decoding invitation: %v", err)
		return
	}

	issuerConnection, err = bouncer.EstablishConnection(invite, 10*time.Second)
	if err != nil {
		util.WriteErrorf(w, "Error requesting invitation from issuer: %v", err)
		return
	}

	d, _ := json.MarshalIndent(issuerConnection, " ", " ")
	_ = subjectStore.Put("issuer", d)
	util.WriteSuccess(w, d)
}

func connectToVerifier(w http.ResponseWriter, req *http.Request) {
	b := req.Body
	defer b.Close()

	invite := &didexchange.Invitation{}
	err := json.NewDecoder(b).Decode(invite)
	if err != nil {
		util.WriteErrorf(w, "Error decoding invitation: %v", err)
		return
	}

	verifierConnection, err = bouncer.EstablishConnection(invite, 10*time.Second)
	if err != nil {
		util.WriteErrorf(w, "Error requesting invitation from verifier: %v", err)
		return
	}

	d, _ := json.MarshalIndent(verifierConnection, " ", " ")
	_ = subjectStore.Put("verifier", d)

	util.WriteSuccess(w, d)
}

func createAriesContext() {
	wsinbound := "172.16.1.1:3001"

	genesis, err := os.Open("./deploy/canis-chart/indy/genesis.txn")
	if err != nil {
		log.Fatalln("unable to open genesis file", err)
	}
	vdrclient, err = vdr.New(genesis)
	if err != nil {
		log.Fatalln("unable to connect to indy vdr", err)
	}

	storeProv := mongodbstore.NewProvider("mongodb://172.17.0.1:27017", mongodbstore.WithDBPrefix("subject"))
	subjectStore, _ = storeProv.OpenStore("connections")
	indyVDRI, err := indy.New("sov", indy.WithIndyClient(vdrclient))
	if err != nil {
		log.Fatalln("unable to create aries indy vdr", err)
	}

	ar, err := aries.New(
		aries.WithStoreProvider(storeProv),
		defaults.WithInboundWSAddr(wsinbound, fmt.Sprintf("ws://%s", wsinbound), "", ""),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithVDRI(indyVDRI),
		aries.WithProtocols(newIssueCredentialSvc()),
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
		log.Fatalln("could not create bouncer", err)
	}

	credHandler, err = NewCredHandler(ctx)
	if err != nil {
		log.Fatalln("unable to create cred handler", err)
	}

	sup, err := credential.New(credHandler)
	if err != nil {
		log.Fatalln("unable to create credential supervisor for subject", err)
	}
	err = sup.Start(credHandler)
	if err != nil {
		log.Fatalln(err, "unable to start credential supervisor for subject")
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

	prover, err = ursa.NewProver(ctx)
	if err != nil {
		log.Fatalln("unable to create Ursa prover")
	}

	d, err := subjectStore.Get("issuer")
	if err == nil {
		issuerConnection = &didexchange.Connection{}
		err = json.Unmarshal(d, issuerConnection)
		if err != nil {
			log.Fatalln("issuer conneciton stored but not valid")
		}
	}

	d, err = subjectStore.Get("verifier")
	if err == nil {
		verifierConnection = &didexchange.Connection{}
		err = json.Unmarshal(d, verifierConnection)
		if err != nil {
			log.Fatalln("verifier conneciton stored but not valid")
		}
	}

}

func newIssueCredentialSvc() api.ProtocolSvcCreator {
	return func(prv api.Provider) (dispatcher.ProtocolService, error) {
		service, err := icprotocol.New(prv)
		if err != nil {
			return nil, err
		}

		// sets default middleware to the service
		// service.Use(mdissuecredential.SaveCredentials(prv))

		return service, nil
	}
}

type credentialHandler struct {
	credcl *issuecredential.Client
	store  vstore.Store
}

func NewCredHandler(ctx *ariescontext.Provider) (*credentialHandler, error) {

	credcl, err := issuecredential.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	vc, err := vstore.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create credential store in steward init")
	}

	a := &credentialHandler{
		credcl: credcl,
		store:  vc,
	}

	return a, nil
}

func (r *credentialHandler) GetCredentialClient() (*issuecredential.Client, error) {
	return r.credcl, nil
}

func (r *credentialHandler) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {

}

func (r *credentialHandler) OfferCredentialMsg(e service.DIDCommAction, d *icprotocol.OfferCredential) {
	fmt.Printf("Credential \"%s\" Offered, accept? (Y/n)\n", d.Comment)
	//b := bufio.NewReader(os.Stdin)
	//answer, err := b.ReadString('\n')
	//if err != nil {
	//	log.Fatalln("error reading stdin")
	//}

	answer := "Y"
	if strings.HasPrefix(strings.ToUpper(answer), "Y") {
		msID, err := prover.CreateMasterSecret(e.Message.ID())
		if err != nil {
			log.Println("error creating master secret", err)
			return
		}

		offer := &ursa.CredentialOffer{}
		bits, _ := d.OffersAttach[0].Data.Fetch()
		err = json.Unmarshal(bits, offer)
		if err != nil {
			log.Println("extract offer from protocol message", err)
			return
		}

		dd, _ := json.MarshalIndent(offer, " ", " ")
		fmt.Println(string(dd))

		rply, err := vdrclient.GetCredDef(offer.CredDefID)
		if err != nil {
			log.Println("unable to retrieve cred def from ledger", err)
			return
		}

		credDef := &vdr.ClaimDefData{}
		err = credDef.UnmarshalReadReply(rply)
		if err != nil {
			log.Println("unable to marshal get cred def from ledger", err)
			return
		}

		credReq, credReqMeta, err := prover.CreateCredentialRequest(issuerConnection.MyDID, credDef, offer, msID)
		if err != nil {
			log.Println("unable to create ursa credential request", err)
			return
		}

		x, _ := json.MarshalIndent(credReqMeta, " ", " ")
		fmt.Println(string(x))

		b, err := json.Marshal(credReq)
		if err != nil {
			log.Println(err, "unexpect error marshalling offer into JSON")
			return
		}

		msg := &icprotocol.RequestCredential{
			Type:    icprotocol.RequestCredentialMsgType,
			Comment: d.Comment,
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					Base64: base64.StdEncoding.EncodeToString(b),
				}},
			},
		}
		e.Continue(icprotocol.WithRequestCredential(msg))
	} else {
		e.Stop(errors.New("not accepted"))
	}

}

func (r *credentialHandler) IssueCredentialMsg(e service.DIDCommAction, d *icprotocol.IssueCredential) {
	fmt.Printf("Credential \"%s\" issued, accept? (Y/n)\n", d.Comment)
	//b := bufio.NewReader(os.Stdin)
	//answer, err := b.ReadString('\n')
	//if err != nil {
	//	log.Fatalln("error reading stdin")
	//}

	answer := "Y"
	if strings.HasPrefix(strings.ToUpper(answer), "Y") {
		//TODO: @m00sey you'll probably ask the user for the name here, instead of comment.
		thid, _ := e.Message.ThreadID()
		err := r.credcl.AcceptCredential(thid)
		if err != nil {
			log.Println("Error accepting credential", err)
			return
		}
		log.Printf("%s Accepted\n", d.Comment)
	}

}

func (r *credentialHandler) RequestCredentialMsg(_ service.DIDCommAction, _ *icprotocol.RequestCredential) {

}

func (r *credentialHandler) GetCredentials() ([]*verifiable.Credential, error) {
	creds, err := r.store.GetCredentials()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load credentials")
	}

	var out []*verifiable.Credential
	for _, cr := range creds {
		fmt.Println(cr.ID, ":", cr.Name)
		cred, err := r.store.GetCredential(cr.ID)
		if err != nil {
			log.Println("error getting cred", cr.ID)
			continue
		}
		out = append(out, cred)
	}

	return out, nil
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

func NewProofHandler(ctx *ariescontext.Provider) (*proofHandler, error) {

	ppcl, err := ppclient.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	vc, err := vstore.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create credential store in steward init")
	}

	a := &proofHandler{
		ppcl:  ppcl,
		store: vc,
	}

	return a, nil
}

type proofHandler struct {
	ppcl  *ppclient.Client
	store vstore.Store
}

func (r *proofHandler) ProposePresentationMsg(_ service.DIDCommAction, _ *presentproof.ProposePresentation) {

}

func (r *proofHandler) RequestPresentationMsg(e service.DIDCommAction, req *presentproof.RequestPresentation) {
	fmt.Println("I have received the following:")
	d, _ := json.MarshalIndent(req, " ", " ")
	fmt.Println(string(d))

	vp := &verifiable.Presentation{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1"},
		ID:     "urn:uuid:3978344f-8596-4c3a-a978-8fcaba3903c",
		Type:   []string{"VerifiablePresentation", "Clr"},
		Holder: verifierConnection.MyDID,
	}

	doc, err := ctx.VDRIRegistry().Resolve(verifierConnection.MyDID)
	if err != nil {
		log.Fatalln("unable to load my did doc")
	}

	signer, err := newCryptoSigner(doc.PublicKey[0].ID[1:])
	if err != nil {
		log.Fatalln(err)
	}

	vc, err := r.store.GetCredential("http://example.edu/credentials/1872")
	if err != nil {
		log.Fatalln("unable to get credential", err)
	}

	err = vp.SetCredentials(vc)
	if err != nil {
		log.Fatalln("unable to set credentials on the presentation", err)
	}

	sigSuite := ed25519signature2018.New(
		suite.WithSigner(signer),
		suite.WithVerifier(ed25519signature2018.NewPublicKeyVerifier()))
	ldpContext := &verifiable.LinkedDataProofContext{
		SignatureType:           "Ed25519Signature2018",
		SignatureRepresentation: verifiable.SignatureProofValue,
		Suite:                   sigSuite,
		VerificationMethod:      fmt.Sprintf("%s%s", verifierConnection.MyDID, doc.PublicKey[0].ID),
	}

	err = vp.AddLinkedDataProof(ldpContext)
	if err != nil {
		log.Fatalln("error adding linked data proof", err)
	}

	pres := &ppclient.Presentation{
		PresentationsAttach: []decorator.Attachment{
			{
				Data: decorator.AttachmentData{
					JSON: vp,
				},
			},
		},
	}
	e.Continue(ppclient.WithPresentation(pres))

}

func (r *proofHandler) PresentationMsg(_ service.DIDCommAction, _ *presentproof.Presentation) {
	panic("implement me")
}

func (r *proofHandler) PresentationPreviewMsg(_ service.DIDCommAction, _ *presentproof.Presentation) {
	panic("implement me")
}
