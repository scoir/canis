package verifier

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
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
	CredDefID() string
}

func (r *proofHandler) ProposePresentationMsg(e service.DIDCommAction, d *ppprotocol.ProposePresentation) {
	panic("implement me")
}

// set of requested attributes
// "<attr_referent>": <attr_info>,
type requestedAttributes []map[string]interface{}

// set of requested predicates
// "<predicate_referent>": <predicate_info>
type requestedPredicates []map[string]interface{}

func (r *proofHandler) RequestPresentationMsg(e service.DIDCommAction, d *ppprotocol.RequestPresentation) {
	//props := e.Properties.(prop)
	//myDID := props.MyDID()
	//theirDID := props.TheirDID()
	//credDefID := props.CredDefID()
	//
	////look up cred def for values?
	//e.Properties

	// look up schema to get values to be revealed?
	//r.store.GetAgentByPublicDID(theirDID)
	//
	////parse attributes, build attachments
	//for k, v := d.RequestPresentationAttach
	//
	//r.registry.RequestPresentation(theirDID, myDID, credDefID)
}

func (r *proofHandler) PresentationMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	panic("implement me")
}

func (r *proofHandler) PresentationPreviewMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	panic("implement me")
}
