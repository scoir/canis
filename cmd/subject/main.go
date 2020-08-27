package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/store/verifiable"
	"github.com/pkg/errors"
	goji "goji.io"
	"goji.io/pat"

	"github.com/scoir/canis/pkg/aries/storage/mongodb/store"
	"github.com/scoir/canis/pkg/credential"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/util"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var issuerConnection *didexchange.Connection
var verifierConnection *didexchange.Connection

func main() {
	arieslog.SetLevel("aries-framework/out-of-band/service", arieslog.CRITICAL)
	arieslog.SetLevel("aries-framework/ws", arieslog.CRITICAL)
	//arieslog.SetLevel("aries-framework/did-exchange/service", arieslog.DEBUG)
	arieslog.SetLevel("aries-framework/issuecredential/service", arieslog.DEBUG)
	createAriesContext()
	listen()
}

func listen() {
	mux := goji.NewMux()
	mux.Handle(pat.Post("/connect-to-issuer"), http.HandlerFunc(connectToIssuer))
	mux.Handle(pat.Post("/connect-to-verifier"), http.HandlerFunc(connectToVerifier))

	u := "0.0.0.0:3002"
	log.Println("subject listening on", u)
	err := http.ListenAndServe(u, mux)
	log.Fatalln("subject no longer listening", err)
}

func connectToIssuer(w http.ResponseWriter, _ *http.Request) {
	resp, err := http.Get("http://0.0.0.0:2002/invitation/subject")
	if err != nil {
		util.WriteErrorf(w, "Error requesting invitation from issuer: %v", err)
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

	h, err := NewCredHandler(ctx)
	if err != nil {
		log.Fatalln("unable to create cred handler")
	}

	sup, err := credential.New(h)
	if err != nil {
		log.Fatalln("unable to create credential supervisor for subject", err)
	}
	err = sup.Start(h)
	if err != nil {
		log.Fatalln(err, "unable to start credential supervisor for subject")
	}

}

type credentialHandler struct {
	credcl *issuecredential.Client
	store  verifiable.Store
}

func NewCredHandler(ctx *ariescontext.Provider) (*credentialHandler, error) {

	credcl, err := issuecredential.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	vc, err := verifiable.New(ctx)
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
	b := bufio.NewReader(os.Stdin)
	answer, err := b.ReadString('\n')
	if err != nil {
		log.Fatalln("error reading stdin")
	}

	if strings.HasPrefix(strings.ToUpper(answer), "Y") {
		err := r.credcl.AcceptOffer(e.Message.ID())
		if err != nil {
			log.Println("Error accepting offer", err)
		}
	}

}

func (r *credentialHandler) IssueCredentialMsg(e service.DIDCommAction, d *icprotocol.IssueCredential) {
	fmt.Printf("Credential \"%s\" issued, accept? (Y/n)\n", d.Comment)
	b := bufio.NewReader(os.Stdin)
	answer, err := b.ReadString('\n')
	if err != nil {
		log.Fatalln("error reading stdin")
	}

	if strings.HasPrefix(strings.ToUpper(answer), "Y") {
		//TODO: @m00sey you'll probably ask the user for the name here, instead of comment.
		thid, _ := e.Message.ThreadID()
		err := r.credcl.AcceptCredential(thid, d.Comment)
		if err != nil {
			log.Println("Error accepting credential", err)
			return
		}
		log.Printf("%s Accepted\n", d.Comment)
	}

	fmt.Println("Credentials in Wallet")
	creds, err := r.store.GetCredentials()
	if err != nil {
		log.Println("error getting stored credentials", err)
		return
	}

	for _, cr := range creds {
		fmt.Println(cr.ID, ":", cr.Name)
		cred, err := r.store.GetCredential(cr.ID)
		if err != nil {
			log.Println("error getting cred", cr.ID)
			continue
		}

		log.Println("\t", cred.Issued)
		log.Println("\t", cred.Issuer.ID)
	}

}

func (r *credentialHandler) RequestCredentialMsg(_ service.DIDCommAction, _ *icprotocol.RequestCredential) {

}
