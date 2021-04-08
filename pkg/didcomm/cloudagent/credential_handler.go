package cloudagent

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/credential/engine/indy"
	"github.com/scoir/canis/pkg/credential/engine/lds"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
	cursa "github.com/scoir/canis/pkg/ursa"
)

const (
	masterSecretID = "my-master-secret"
)

type credentialHandler struct {
	credcl    *issuecredential.Client
	prover    *cursa.Prover
	vdrclient *vdr.Client
	store     datastore.Store
}

type eventProps interface {
	MyDID() string
	TheirDID() string
}

func NewCredHandler(ctx provider) (*credentialHandler, error) {

	actx, err := ctx.GetAriesContext()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries provider for cloud agent credential handler")
	}

	credcl, err := issuecredential.New(actx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	vdrclient, err := ctx.GetVDRClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get VDR client for cloud agent")
	}

	store, err := ctx.GetDatastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get canis datastore for cloud agent")
	}

	prover, err := cursa.NewProver(actx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Ursa prover")
	}

	a := &credentialHandler{
		credcl:    credcl,
		store:     store,
		vdrclient: vdrclient,
		prover:    prover,
	}

	return a, nil
}

func (r *credentialHandler) GetCredentialClient() (*issuecredential.Client, error) {
	return r.credcl, nil
}

func (r *credentialHandler) OfferCredentialMsg(e service.DIDCommAction, d *icprotocol.OfferCredential) {
	for _, format := range d.Formats {
		for _, attach := range d.OffersAttach {
			if attach.ID == format.AttachID {
				switch format.Format {
				case lds.LinkedDataSignature:
					r.saveLDSOffer(e, d, attach)
				case indy.Indy:
					r.saveIndyOffer(e, d, attach)
				}
				continue
			}
		}
	}
}

func (r *credentialHandler) IssueCredentialMsg(e service.DIDCommAction, d *icprotocol.IssueCredential) {

	for _, format := range d.Formats {
		for _, attach := range d.CredentialsAttach {
			if attach.ID == format.AttachID {
				switch format.Format {
				case lds.LinkedDataSignature:
					r.acceptLDSOffer(e, d, attach)
				case indy.Indy:
					r.acceptIndyOffer(e, d, attach)
				}
				continue
			}
		}
	}

}

func (r *credentialHandler) saveLDSOffer(e service.DIDCommAction, offer *icprotocol.OfferCredential, attachment decorator.Attachment) {
	ep := e.Properties.(eventProps)
	myDID := ep.MyDID()
	theirDID := ep.TheirDID()
	thid, _ := e.Message.ThreadID()

	cloudAgentConnection, err := r.store.GetCloudAgentConnectionForDIDs(myDID, theirDID)
	if err != nil {
		log.Println("uable to load cloud agent connection for LDS offer", err)
		return
	}

	fmt.Println("Accepting credential offer", offer.Comment)
	d, _ := attachment.Data.Fetch()
	cloudAgentCredential := &datastore.CloudAgentCredential{
		ID:           uuid.New().String(),
		CloudAgentID: cloudAgentConnection.CloudAgentID,
		SystemState:  "offered",
		Format:       lds.LinkedDataSignature,
		MyDID:        myDID,
		TheirDID:     theirDID,
		ThreadID:     thid,
		IssuerConnection: &datastore.IDName{
			ID:   cloudAgentConnection.ConnectionID,
			Name: cloudAgentConnection.TheirLabel,
		},
		Offer: &datastore.Offer{
			Comment: offer.Comment,
			Type:    offer.Type,
			Data:    d,
		},
	}

	err = r.store.InsertCloudAgentCredential(cloudAgentCredential)
	if err != nil {
		log.Println("unable to save cloud agent credential", err)
		return
	}

	log.Println("Credential offer saved from", theirDID, "to", myDID)

}

func (r *credentialHandler) saveIndyOffer(e service.DIDCommAction, d *icprotocol.OfferCredential, attachment decorator.Attachment) {
	ep := e.Properties.(eventProps)
	myDID := ep.MyDID()
	theirDID := ep.TheirDID()
	thid, _ := e.Message.ThreadID()

	cloudAgentConnection, err := r.store.GetCloudAgentConnectionForDIDs(myDID, theirDID)
	if err != nil {
		log.Println("uable to load cloud agent connection for Indy offer", err)
		return
	}

	fmt.Println("Accepting credential offer", d.Comment)
	ms, err := r.prover.CreateMasterSecret(masterSecretID)
	if err != nil {
		log.Println("error creating master secret", err)
		return
	}

	offer := &schema.IndyCredentialOffer{}
	bits, _ := attachment.Data.Fetch()
	err = json.Unmarshal(bits, offer)
	if err != nil {
		log.Println("extract offer from protocol message", err)
		return
	}

	rply, err := r.vdrclient.GetCredDef(offer.CredDefID)
	if err != nil {
		log.Println("unable to retrieve cred def from ledger", err)
		return
	}

	credDef := &vdr.ClaimDefData{ID: offer.CredDefID}
	err = credDef.UnmarshalReadReply(rply)
	if err != nil {
		log.Println("unable to marshal get cred def from ledger", err)
		return
	}

	credReq, credReqMeta, err := r.prover.CreateCredentialRequest(myDID, credDef, offer, ms)
	if err != nil {
		log.Println("unable to create ursa credential request", err)
		return
	}

	vals := map[string]interface{}{}
	attrs := make([]icprotocol.Attribute, len(d.CredentialPreview.Attributes))
	for i, a := range d.CredentialPreview.Attributes {
		attrs[i] = icprotocol.Attribute{
			Name:  a.Name,
			Value: a.Value,
		}
		vals[a.Name] = a.Value
	}

	cloudAgentCredential := &datastore.CloudAgentCredential{
		ID:           uuid.New().String(),
		CloudAgentID: cloudAgentConnection.CloudAgentID,
		SystemState:  "offered",
		Format:       indy.Indy,
		MyDID:        myDID,
		TheirDID:     theirDID,
		ThreadID:     thid,
		IssuerConnection: &datastore.IDName{
			ID:   cloudAgentConnection.ConnectionID,
			Name: cloudAgentConnection.TheirLabel,
		},
		CredentialRequest:         credReq,
		CredentialRequestMetadata: credReqMeta,
		Offer: &datastore.Offer{
			Comment: d.Comment,
			Type:    d.Type,
			Preview: attrs,
		},
	}

	err = r.store.InsertCloudAgentCredential(cloudAgentCredential)
	if err != nil {
		log.Println("unable to save cloud agent credential", err)
		return
	}

	log.Println("Credential offer saved from", theirDID, "to", myDID)

}

func (r *credentialHandler) acceptLDSOffer(e service.DIDCommAction, issue *icprotocol.IssueCredential, attachment decorator.Attachment) {
	fmt.Println("Accepting credential")
	thid, _ := e.Message.ThreadID()
	err := r.credcl.AcceptCredential(thid)
	if err != nil {
		log.Println("Error accepting credential", err)
		return
	}

	ep := e.Properties.(eventProps)
	myDID := ep.MyDID()
	theirDID := ep.TheirDID()
	cloudAgentConnection, err := r.store.GetCloudAgentConnectionForDIDs(myDID, theirDID)
	if err != nil {
		log.Println("cloud agent conneciton not found from", theirDID, "to", myDID)
		return
	}

	cloudAgentCredential, err := r.store.GetCloudAgentCredentialFromThread(cloudAgentConnection.CloudAgentID, thid)
	if err != nil {
		log.Println("cloud agent credential", thid, "not found for", cloudAgentConnection.CloudAgentID)
		return
	}

	data, err := attachment.Data.Fetch()
	if err != nil {
		log.Println("error fetching credential", err)
		return
	}

	fmt.Println(string(data))

	cloudAgentCredential.Credential = &datastore.Credential{
		Description: issue.Comment,
		MimeType:    lds.LinkedDataSignature,
		LastModTime: time.Now(),
		Data:        data,
	}
	cloudAgentCredential.SystemState = "issued"

	err = r.store.UpdateCloudAgentCredential(cloudAgentCredential)
	if err != nil {
		log.Println("unable to update issued cloud agent credential", err)
	}

}

func (r *credentialHandler) acceptIndyOffer(e service.DIDCommAction, issue *icprotocol.IssueCredential, attachment decorator.Attachment) {
	fmt.Println("Accepting credential")
	thid, _ := e.Message.ThreadID()
	err := r.credcl.AcceptCredential(thid)
	if err != nil {
		log.Println("Error accepting credential", err)
		return
	}

	ep := e.Properties.(eventProps)
	myDID := ep.MyDID()
	theirDID := ep.TheirDID()
	cloudAgentConnection, err := r.store.GetCloudAgentConnectionForDIDs(myDID, theirDID)
	if err != nil {
		log.Println("cloud agent conneciton not found from", theirDID, "to", myDID)
		return
	}

	cloudAgentCredential, err := r.store.GetCloudAgentCredentialFromThread(cloudAgentConnection.CloudAgentID, thid)
	if err != nil {
		log.Println("cloud agent credential", thid, "not found for", cloudAgentConnection.CloudAgentID)
		return
	}

	data, err := attachment.Data.Fetch()
	if err != nil {
		log.Println("error fetching credential", err)
		return
	}

	credRequestMedataData := cloudAgentCredential.CredentialRequestMetadata
	credRequest := cloudAgentCredential.CredentialRequest

	cred := &schema.IndyCredential{}
	err = json.Unmarshal(data, cred)
	if err != nil {
		log.Println("error decoding credential", err)
		return
	}

	rply, err := r.vdrclient.GetCredDef(credRequest.CredDefID)
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

	ms, err := r.prover.GetMasterSecret(masterSecretID)
	if err != nil {
		log.Println("unable to get master secret", err)
		return
	}

	sig, err := r.prover.ProcessCredentialSignature(cred, credRequest, ms, credRequestMedataData.MasterSecretBlindingData, credDef.PKey())
	if err != nil {
		log.Println("unable to process credential signature", err)
	}

	cred.Signature = []byte(sig)
	data, _ = json.Marshal(cred)

	cloudAgentCredential.Credential = &datastore.Credential{
		Description: issue.Comment,
		MimeType:    indy.Indy,
		LastModTime: time.Now(),
		Data:        data,
	}

	err = r.store.UpdateCloudAgentCredential(cloudAgentCredential)
	if err != nil {
		log.Println("unable to update issued cloud agent credential", err)
	}

}

//No-ops
func (r *credentialHandler) RequestCredentialMsg(_ service.DIDCommAction, _ *icprotocol.RequestCredential) {

}

func (r *credentialHandler) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {

}
