package verifier

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/presentproof"
	"github.com/scoir/canis/pkg/presentproof/engine"
)

type proofHandler struct {
	ctx      *ariescontext.Provider
	store    datastore.Store
	ppsup    *presentproof.Supervisor
	registry engine.PresentationRegistry
}

type prop interface {
	MyDID() string
	TheirDID() string
	PIID() string
}

func (r *proofHandler) ProposePresentationMsg(e service.DIDCommAction, d *ppprotocol.ProposePresentation) {
	log.Println("presentation proposal not implemented")
}

func (r *proofHandler) RequestPresentationMsg(e service.DIDCommAction, d *ppprotocol.RequestPresentation) {
	log.Println("request presentation not implemented")
}

func (r *proofHandler) PresentationMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	props, ok := e.Properties.(prop)
	if !ok {
		log.Println("presentation properties invalid")
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
			err := errors.Errorf("unexpected error verifying %d presentation: (%+v)", i, err)
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

func (r *proofHandler) PresentationPreviewMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	log.Println("presentation preview not implemented")
}
