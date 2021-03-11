package cloudagent

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"

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

func (r *credentialHandler) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {

}

func (r *credentialHandler) OfferCredentialMsg(e service.DIDCommAction, d *icprotocol.OfferCredential) {

	ep := e.Properties.(eventProps)
	myDID := ep.MyDID()
	theirDID := ep.TheirDID()
	thid, _ := e.Message.ThreadID()

	cloudAgentConnection, err := r.store.GetCloudAgentConnectionForDIDs(myDID, theirDID)

	fmt.Println("Accepting credential offer", d.Comment)
	ms, err := r.prover.CreateMasterSecret(masterSecretID)
	if err != nil {
		log.Println("error creating master secret", err)
		return
	}

	offer := &schema.IndyCredentialOffer{}
	bits, _ := d.OffersAttach[0].Data.Fetch()
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
		ID:                        uuid.New().String(),
		CloudAgentID:              cloudAgentConnection.CloudAgentID,
		SystemState:               "offered",
		MyDID:                     myDID,
		TheirDID:                  theirDID,
		ThreadID:                  thid,
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

	//msg := &icprotocol.RequestCredential{
	//	Type:    icprotocol.RequestCredentialMsgType,
	//	Comment: d.Comment,
	//	RequestsAttach: []decorator.Attachment{
	//		{Data: decorator.AttachmentData{
	//			Base64: base64.StdEncoding.EncodeToString(b),
	//		}},
	//	},
	//}

	//e.Continue(icprotocol.WithRequestCredential(msg))
}

func (r *credentialHandler) IssueCredentialMsg(e service.DIDCommAction, d *icprotocol.IssueCredential) {
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

	data, err := d.CredentialsAttach[0].Data.Fetch()
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
		Description: d.Comment,
		MimeType:    d.Formats[0].Format,
		LastModTime: time.Now(),
		Data:        data,
	}

	err = r.store.UpdateCloudAgentCredential(cloudAgentCredential)
	if err != nil {
		log.Println("unable to update issued cloud agent credential", err)
	}

}

func (r *credentialHandler) RequestCredentialMsg(_ service.DIDCommAction, _ *icprotocol.RequestCredential) {

}
