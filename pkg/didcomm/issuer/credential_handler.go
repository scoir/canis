package issuer

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/amqp"
	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/notifier"
)

type CredHandler struct {
	store                 datastore.Store
	registry              engine.CredentialRegistry
	notificationPublisher amqp.Publisher
}

func NewCredentialHandler(store datastore.Store, reg engine.CredentialRegistry) *CredHandler {
	handler := &CredHandler{
		store:    store,
		registry: reg,
	}

	return handler
}

type prop interface {
	MyDID() string
	TheirDID() string
}

func (r *CredHandler) ProposeCredentialMsg(e service.DIDCommAction, proposal *icprotocol.ProposeCredential) {
	thid, _ := e.Message.ThreadID()
	offer, err := r.store.FindCredentialByProtocolID(thid)
	if err == nil {
		e.Stop(errors.New("negociating not currently supported"))
		err = r.store.DeleteCredentialByOffer(offer.ProtocolID)
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

	cred := &datastore.IssuedCredential{
		AgentName:         agent.Name,
		MyDID:             myDID,
		TheirDID:          theirDID,
		SchemaName:        schema.Name,
		ExternalSubjectID: ac.ExternalID,
		ProtocolID:        thid,
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

func (r *CredHandler) OfferCredentialMsg(_ service.DIDCommAction, _ *icprotocol.OfferCredential) {
	//NO-OP - this is a Holder State
}

func (r *CredHandler) IssueCredentialMsg(_ service.DIDCommAction, _ *icprotocol.IssueCredential) {
	//NO-OP - this is a Holder State
}

func (r *CredHandler) RequestCredentialMsg(e service.DIDCommAction, request *icprotocol.RequestCredential) {

	if len(request.RequestsAttach) != 1 {
		log.Println("only one credential is supported at a time")
		e.Stop(errors.New("only one credential is supported at a time"))
		return
	}

	thid, _ := e.Message.ThreadID()
	cred, err := r.store.FindCredentialByProtocolID(thid)
	if err != nil {
		log.Printf("unable to find cred with ID %s: (%v)\n", thid, err)
		e.Stop(errors.Errorf("unable to find cred with ID %s: (%v)", thid, err))
		return
	}

	_, err = r.store.GetAgent(cred.AgentName)
	if err != nil {
		log.Println("unable to find agent for credential request", err)
		e.Stop(errors.Errorf("unable to find agent for credential request: %v", err))
		return
	}

	schema, err := r.store.GetSchema(cred.SchemaName)
	if err != nil {
		log.Printf("unable to find schema with ID %s: (%v)\n", cred.SchemaName, err)
		e.Stop(errors.Errorf("unable to find schema with ID %s: (%v)", cred.SchemaName, err))
		return
	}

	did, err := r.store.GetDID(cred.MyDID)
	if err != nil {
		log.Printf("unable to find DID with ID %s: (%v)\n", cred.MyDID, err)
		e.Stop(errors.Errorf("unable to find DID with ID %s: (%v)", cred.MyDID, err))
		return
	}

	values := map[string]interface{}{}
	for _, attr := range cred.Offer.Preview {
		values[attr.Name] = attr.Value
	}

	requestAttachment := request.RequestsAttach[0]
	attachmentData, err := r.registry.IssueCredential(did, schema, cred.RegistryOfferID,
		requestAttachment.Data, values)
	if err != nil {
		msg := fmt.Sprintf("registry error creating credential: %v", err)
		fmt.Println(msg)
		e.Stop(errors.New(msg))
		return
	}

	credentialAttachment := decorator.Attachment{
		ID:          uuid.New().String(),
		MimeType:    "application/json",
		LastModTime: time.Now(),
		Data:        *attachmentData,
	}

	msg := &icprotocol.IssueCredential{
		Comment: cred.Offer.Comment,
		Formats: []icprotocol.Format{
			{
				AttachID: credentialAttachment.ID,
				Format:   "hlindy-zkp-v1.0",
			},
		},
		CredentialsAttach: []decorator.Attachment{credentialAttachment},
	}

	d, err := attachmentData.Fetch()
	if err != nil {
		e.Stop(errors.Errorf("unable to fetch attachment: %v", err))
		return
	}

	dscred := &datastore.Credential{
		ID:          credentialAttachment.ID,
		MimeType:    credentialAttachment.MimeType,
		LastModTime: credentialAttachment.LastModTime,
		Data:        d,
	}

	cred.Credential = dscred

	err = r.store.UpdateCredential(cred)
	if err != nil {
		e.Stop(errors.Errorf("unexpected error updating issued credential %s: %v", cred.ID, err))
		return
	}

	e.Continue(icprotocol.WithIssueCredential(msg))
}

func (r *CredHandler) publishProposalReceived(agent *datastore.Agent, externalID string, schema *datastore.Schema,
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

func (r *CredHandler) publishEvent(evt interface{}) error {

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
