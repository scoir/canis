package issuer

import (
	"fmt"
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/datastore"
)

type credHandler struct {
	ctx     *ariescontext.Provider
	credsup *credential.Supervisor
	store   datastore.Store
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
	props := e.Properties.(prop)
	myDID := props.MyDID()

	if len(request.RequestsAttach) != 1 {
		log.Println("invalid request attachments")
		return
	}

	agent, err := r.store.GetAgentByPublicDID(myDID)
	if err != nil {
		log.Println("unable to find agent for credential request", err)
		return
	}

	offer, err := r.store.FindOffer(agent.ID, thid)
	if err != nil {
		log.Printf("unable to find offer with ID %s: (%v)\n", thid, err)
		return
	}

	//TODO:  Somehow verify the request against the original offer
	fmt.Printf("offerID: %s, threadID: %s\n", offer.OfferID, thid)
	msg := &icprotocol.IssueCredential{
		Comment: offer.Offer.Comment,
		CredentialsAttach: []decorator.Attachment{
			{Data: decorator.AttachmentData{JSON: "insert indy magic here"}},
		},
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
