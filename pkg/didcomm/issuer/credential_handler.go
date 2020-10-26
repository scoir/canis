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

func (r *credHandler) ProposeCredentialMsg(e service.DIDCommAction, proposal *icprotocol.ProposeCredential) {
	thid, _ := e.Message.ThreadID()

	offer, err := r.store.FindOffer(thid)
	if err == nil {
		_, err = r.store.GetAgent(offer.AgentID)
		if err != nil {
			log.Println("unable to find agent for credential request", err)
			return
		}

		//TODO: implement negociation from previous offer
		return
	}

	//proposal.CredentialProposal
	////No existing offer, a proposal for a new credential offer
	//proposal.Formats[0].Format
	//proposal.Formats[0].AttachID
	//
	////FIND SCHEMA WITH THIS FORMAT AND DO SOMETHING WITH THE BASE64 BELOW...
	//proposal.FilterAttach[0].ID
	//proposal.FilterAttach[0].Data.Base64
	//
	//registryOfferID, attachment, err := r.registry.CreateCredentialOffer(agent.PublicDID, ac.TheirDID, schema, vals)
	//if err != nil {
	//	return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error creating credential offer: %v", err))
	//}

}

func (r *credHandler) OfferCredentialMsg(_ service.DIDCommAction, _ *icprotocol.OfferCredential) {
	//NO-OP - this is a Holder State
}

func (r *credHandler) IssueCredentialMsg(_ service.DIDCommAction, _ *icprotocol.IssueCredential) {
	//NO-OP - this is a Holder State
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
	for _, attr := range offer.Offer.Preview {
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
	mon.WatchThread(thid, r.CredentialAccepted(offer.OfferID), r.CredentialError)
	e.Continue(icprotocol.WithIssueCredential(msg))
}

func (r *credHandler) CredentialAccepted(id string) func(threadID string, ack *model.Ack) {

	return func(threadID string, ack *model.Ack) {
		//TODO: find the offer and update the status and send notification!!
		fmt.Printf("Transcript Accepted: %s", id)
	}
}

func (r *credHandler) CredentialError(threadID string, err error) {
	//TODO: find the offer and update the status and send notification!!
	log.Println("step 1... failed!", threadID, err)
}
