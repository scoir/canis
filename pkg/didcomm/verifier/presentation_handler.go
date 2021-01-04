package verifier

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/presentproof/engine"
)

func NewProofHandler(store datastore.Store, reg engine.PresentationRegistry) *ProofHandler {
	return &ProofHandler{
		store:    store,
		registry: reg,
	}
}

type ProofHandler struct {
	store    datastore.Store
	registry engine.PresentationRegistry
}

type prop interface {
	MyDID() string
	TheirDID() string
	PIID() string
}

func (r *ProofHandler) ProposePresentationMsg(e service.DIDCommAction, _ *ppprotocol.ProposePresentation) {
	err := errors.New("presentation proposal not implemented")
	e.Stop(err)
}

func (r *ProofHandler) RequestPresentationMsg(e service.DIDCommAction, _ *ppprotocol.RequestPresentation) {
	err := errors.New("request presentation not implemented")
	e.Stop(err)
}

func (r *ProofHandler) PresentationMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	props, ok := e.Properties.(prop)
	if !ok {
		err := errors.New("presentation properties invalid")
		e.Stop(err)
		return
	}

	myDID := props.MyDID()
	theirDID := props.TheirDID()
	piid := props.PIID()

	pr, err := r.store.GetPresentationRequest(piid)
	if err != nil {
		err := errors.Errorf("unable to find presentation request %s", piid)
		log.Println(err)
		e.Stop(err)
		return
	}

	verified := make([]*datastore.Presentation, len(d.PresentationsAttach))
	for i, format := range d.Formats {
		presentationsAttach, ok := getAttachment(format.AttachID, d.PresentationsAttach)
		if !ok {
			err := errors.Errorf("presentations and formats do not match %d", i)
			log.Println(err)
			e.Stop(err)
			return
		}

		proofData, err := presentationsAttach.Data.Fetch()
		if err != nil {
			err := errors.Errorf("unable to fetch presentation data from proof %d: (%v)", i, err)
			log.Println(err)
			e.Stop(err)
			return
		}

		err = r.registry.Verify(format.Format, proofData, pr.Data, theirDID, myDID)
		if err != nil {
			err := errors.Errorf("unexpected error verifying %d presentation: (%v)", i, err)
			log.Println(err)
			e.Stop(err)
			return
		}

		presentation := &datastore.Presentation{
			TheirDID: theirDID,
			MyDID:    myDID,
			Format:   format.Format,
			Data:     proofData,
		}
		verified[i] = presentation

	}

	for _, v := range verified {

		_, err := r.store.InsertPresentation(v)
		if err != nil {
			err := errors.Errorf("unexpected error saving verified presention: (%v)", err)
			e.Stop(err)
			return
		}
	}

	e.Continue(piid) //WithFriendlyNames?
}

func getAttachment(attachID string, attach []decorator.Attachment) (*decorator.Attachment, bool) {
	for _, attachment := range attach {
		if attachment.ID == attachID {
			return &attachment, true
		}
	}
	return nil, false
}

func (r *ProofHandler) PresentationPreviewMsg(e service.DIDCommAction, _ *ppprotocol.Presentation) {
	err := errors.New("presentation preview not implemented")
	e.Stop(err)
}
