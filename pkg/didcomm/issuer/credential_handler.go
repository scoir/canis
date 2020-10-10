package issuer

import (
	"errors"
	"fmt"
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
)

type credHandler struct {
	ctx      *ariescontext.Provider
	store    datastore.Store
	credsup  *credential.Supervisor
	registry engine.CredentialRegistry
}

type prop interface {
	MyDID() string
	TheirDID() string
}

func (r *credHandler) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {
	panic("implement me")
}

func (r *credHandler) OfferCredentialMsg(_ service.DIDCommAction, _ *icprotocol.OfferCredential) {
	panic("implement me")
}

func (r *credHandler) IssueCredentialMsg(_ service.DIDCommAction, _ *icprotocol.IssueCredential) {
	panic("implement me")
}

func (r *credHandler) RequestCredentialMsg(e service.DIDCommAction, request *icprotocol.RequestCredential) {
	thid, _ := e.Message.ThreadID()

	offer, err := r.store.FindOffer(thid)
	if err != nil {
		log.Printf("unable to find offer with ID %s: (%v)\n", thid, err)
		return
	}

	_, err = r.store.GetAgent(offer.AgentID)
	if err != nil {
		log.Println("unable to find agent for credential request", err)
		return
	}

	schema, err := r.store.GetSchema(offer.SchemaID)
	if err != nil {
		log.Printf("unable to find schema with ID %s: (%v)\n", offer.SchemaID, err)
		return
	}

	did, err := r.store.GetDID(offer.MyDID)
	if err != nil {
		log.Printf("unable to find DID with ID %s: (%v)\n", offer.MyDID, err)
		return
	}

	values := map[string]interface{}{}
	for _, attr := range offer.Offer.Attributes {
		//TODO:  do we have to consider mime-type here and convert?
		values[attr.Name] = attr.Value
	}

	var credentialAttachments []decorator.Attachment
	for _, requestAttachment := range request.RequestsAttach {

		attachmentData, err := r.registry.IssueCredential(did, schema, offer.RegistryOfferID,
			requestAttachment.Data, values)
		if err != nil {
			log.Println("registry error creating credential", err)
			continue
		}

		credentialAttachments = append(credentialAttachments, decorator.Attachment{Data: *attachmentData})
	}

	if len(credentialAttachments) == 0 {
		log.Println("no credentials to issue")
		e.Stop(errors.New("no credentials to issue"))
		return
	}

	//TODO:  Somehow verify the request against the original offer
	fmt.Printf("offerID: %s, threadID: %s\n", offer.OfferID, thid)
	msg := &icprotocol.IssueCredential{
		Comment:           offer.Offer.Comment,
		CredentialsAttach: credentialAttachments,
	}

	//TODO:  Shouldn't this be built into the Supervisor??
	log.Println("setting up monitoring for", thid)
	mon := credential.NewMonitor(r.credsup)
	mon.WatchThread(thid, r.TranscriptAccepted(offer.OfferID), r.CredentialError)

	e.Continue(icprotocol.WithIssueCredential(msg))
}

func (r *credHandler) TranscriptAccepted(id string) func(threadID string, ack *model.Ack) {

	return func(threadID string, ack *model.Ack) {
		//TODO: find the offer and update the status!!
		fmt.Printf("Transcript Accepted: %s", id)
	}
}

func (r *credHandler) CredentialError(threadID string, err error) {
	//TODO: find the offer and update the status!!
	log.Println("step 1... failed!", threadID, err)
}
