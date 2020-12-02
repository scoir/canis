package issuer

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/amqp"
	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/notifier"
)

type credHandler struct {
	ctx                   *ariescontext.Provider
	store                 datastore.Store
	credsup               *credential.Supervisor
	registry              engine.CredentialRegistry
	notificationPublisher amqp.Publisher
}

type prop interface {
	MyDID() string
	TheirDID() string
}

func (r *credHandler) ProposeCredentialMsg(e service.DIDCommAction, proposal *icprotocol.ProposeCredential) {
	thid, _ := e.Message.ThreadID()
	offer, err := r.store.FindOffer(thid)
	if err == nil {
		e.Stop(errors.New("negociating not currently supported"))
		err = r.store.DeleteOffer(offer.ThreadID)
		if err != nil {
			log.Println("unable to delete offer in negociation", err)
			return
		}
		return
	}

	props := e.Properties.(prop)
	myDID := props.MyDID()
	theirDID := props.TheirDID()

	agent, err := r.store.GetAgentByPublicDID(myDID)
	if err != nil {
		log.Println("unable to find agent for proposed credential", err)
		return
	}

	ac, err := r.store.GetAgentConnectionForDID(agent, theirDID)
	if err != nil {
		log.Println("proposed credential from a DID that is not a connection", err)
		return
	}

	attachments := map[string]decorator.Attachment{}
	for _, attach := range proposal.FilterAttach {
		attachments[attach.ID] = attach
	}

	var schema *datastore.Schema
	for _, format := range proposal.Formats {
		attach := attachments[format.AttachID]

		data, err := attach.Data.Fetch()
		if err != nil {
			log.Println("unable to fetch data from proposal", err)
			continue
		}

		schemaID, err := r.registry.GetSchemaForProposal(format.Format, data)
		if err != nil || !agent.CanIssue(schemaID) {
			log.Printf("invalid request for schema %s against agent %s", schemaID, agent.Name)
			continue
		}

		schema, err = r.store.GetSchema(schemaID)
		if err != nil {
			log.Println("registry returned invalid schema ID", err)
			continue
		}

		//Once we find a valid attachment/format, get out.
		break
	}

	if schema == nil {
		log.Println("no supported schema found")
		return
	}

	cred := &datastore.Credential{
		AgentName:         agent.Name,
		MyDID:             myDID,
		TheirDID:          theirDID,
		SchemaName:        schema.Name,
		ExternalSubjectID: ac.ExternalID,
		ThreadID:          thid,
		SystemState:       "propsed",
	}

	_, err = r.store.InsertCredential(cred)
	if err != nil {
		log.Printf("unexpected error saving credential: %v", err)
		return
	}

	err = r.publishProposalReceived(agent, ac.ExternalID, schema, proposal.CredentialProposal)
	if err != nil {
		log.Printf("unexpected error publishing credential proposal webhook: %v", err)
	}

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

	_, err = r.store.GetAgent(offer.AgentName)
	if err != nil {
		log.Println("unable to find agent for credential request", err)
		return
	}

	schema, err := r.store.GetSchema(offer.SchemaName)
	if err != nil {
		log.Printf("unable to find schema with ID %s: (%v)\n", offer.SchemaName, err)
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
	fmt.Printf("offerID: %s, threadID: %s\n", offer.ThreadID, thid)
	msg := &icprotocol.IssueCredential{
		Comment: offer.Offer.Comment,
		Formats: []icprotocol.Format{
			{
				AttachID: credentialAttachments[0].ID,
				Format:   "hlindy-zkp-v1.0",
			},
		},
		CredentialsAttach: credentialAttachments,
	}

	//TODO:  Shouldn't this be built into the Supervisor??
	log.Println("setting up monitoring for", thid)
	mon := credential.NewMonitor(r.credsup)
	mon.WatchThread(thid, r.CredentialAccepted(offer.ThreadID), r.CredentialError)
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

func (r *credHandler) publishProposalReceived(agent *datastore.Agent, externalID string, schema *datastore.Schema,
	proposal icprotocol.PreviewCredential) error {

	evt := &notifier.Notification{
		Topic: CredentialTopic,
		Event: ProposedEvent,
		EventData: CredentialProposalEvent{
			AgentID:    agent.ID,
			ExternalID: externalID,
			Schema:     schema,
			Proposal:   proposal,
		},
	}

	return r.publishEvent(evt)
}

func (r *credHandler) publishEvent(evt interface{}) error {

	message, err := json.Marshal(evt)
	if err != nil {
		return errors.Wrap(err, "unexpected error marshalling did accepted event")
	}

	fmt.Println(string(message))
	err = r.notificationPublisher.Publish(message, "application/json")
	if err != nil {
		return errors.Wrap(err, "unable to publish credential event")
	}

	return nil
}
