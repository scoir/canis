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

func (r *proofHandler) RequestPresentationMsg(e service.DIDCommAction, d *ppprotocol.RequestPresentation) {
	panic("implement me")
}

func (r *proofHandler) PresentationMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	panic("implement me")
}

func (r *proofHandler) PresentationPreviewMsg(e service.DIDCommAction, d *ppprotocol.Presentation) {
	panic("implement me")
}
