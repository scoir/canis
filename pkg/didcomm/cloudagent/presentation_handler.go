package cloudagent

import (
	"log"

	"github.com/google/uuid"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	vstore "github.com/hyperledger/aries-framework-go/pkg/store/verifiable"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

type evtProps interface {
	MyDID() string
	TheirDID() string
	PIID() string
	Err() error
}

type PresentationHandler struct {
	ppcl    *ppclient.Client
	vcstore vstore.Store
	kms     kms.KeyManager
	vdr     vdriapi.Registry
	store   datastore.Store
}

func NewProofHandler(ctx *ariescontext.Provider) (*PresentationHandler, error) {

	ppcl, err := ppclient.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	vc, err := vstore.New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create credential store in steward init")
	}

	return &PresentationHandler{
		ppcl:    ppcl,
		vcstore: vc,
		kms:     ctx.KMS(),
		vdr:     ctx.VDRIRegistry(),
	}, nil
}

func (r *PresentationHandler) ProposePresentationMsg(_ service.DIDCommAction, _ *presentproof.ProposePresentation) {

}

func (r *PresentationHandler) RequestPresentationMsg(e service.DIDCommAction, req *presentproof.RequestPresentation) {
	props := e.Properties.(evtProps)

	cloudAgent, err := r.store.GetCloudAgentForDID(props.MyDID())
	if err != nil {
		log.Println("unable to get cloud agent for presentation request", err)
		return
	}

	pr := &datastore.CloudAgentProofRequest{
		ID:                  uuid.New().String(),
		CloudAgentID:        cloudAgent.ID,
		SystemState:         "requested",
		MyDID:               props.MyDID(),
		TheirDID:            props.TheirDID(),
		ThreadID:            props.PIID(),
		RequestPresentation: req,
	}

	err = r.store.InsertCloudAgentProofRequest(pr)
	if err != nil {
		log.Println("unexpected error saving proof request for cloud agent", err)
	}
}

func (r *PresentationHandler) PresentationMsg(_ service.DIDCommAction, _ *presentproof.Presentation) {
	panic("implement me")
}

func (r *PresentationHandler) PresentationPreviewMsg(_ service.DIDCommAction, _ *presentproof.Presentation) {
	panic("implement me")
}
