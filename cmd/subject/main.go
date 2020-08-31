package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	vstore "github.com/hyperledger/aries-framework-go/pkg/store/verifiable"
	"github.com/pkg/errors"
	goji "goji.io"
	"goji.io/pat"

	"github.com/scoir/canis/pkg/aries/storage/mongodb/store"
	"github.com/scoir/canis/pkg/credential"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	canisproof "github.com/scoir/canis/pkg/presentproof"
	"github.com/scoir/canis/pkg/util"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var issuerConnection *didexchange.Connection
var verifierConnection *didexchange.Connection
var credHandler *credentialHandler
var pofHandler *proofHandler

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

func connectToIssuer(w http.ResponseWriter, _ *http.Request) {
	resp, err := http.Get("http://0.0.0.0:2002/invitation/subject")
	if err != nil {
		util.WriteErrorf(w, "Error requesting invitation from issuer: %v", err)
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	invite := &didexchange.Invitation{}
	err = json.NewDecoder(bytes.NewReader(b)).Decode(invite)
	if err != nil {
		util.WriteErrorf(w, "Error decoding invitation: %v, s", err, string(b))
		return
	}

	issuerConnection, err = bouncer.EstablishConnection(invite, 10*time.Second)
	if err != nil {
		util.WriteErrorf(w, "Error requesting invitation from issuer: %v", err)
		return
	}

	d, _ := json.MarshalIndent(issuerConnection, " ", " ")
	util.WriteSuccess(w, d)
}

func connectToVerifier(w http.ResponseWriter, _ *http.Request) {
	resp, err := http.Get("http://0.0.0.0:4002/invitation/subject")
	if err != nil {
		util.WriteErrorf(w, "Error requesting invitation from verifier: %v", err)
		return
	}
	b := resp.Body
	defer b.Close()

	invite := &didexchange.Invitation{}
	err = json.NewDecoder(b).Decode(invite)
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
	util.WriteSuccess(w, d)
}

func createAriesContext() {
	wsinbound := "0.0.0.0:3001"

	ar, err := aries.New(
		aries.WithStoreProvider(store.NewProvider("mongodb://172.17.0.1:27017", "subject")),
		defaults.WithInboundWSAddr(wsinbound, wsinbound, "", ""),
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
		err := r.credcl.AcceptOffer(e.Message.ID())
		if err != nil {
			log.Println("Error accepting offer", err)
		}
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
