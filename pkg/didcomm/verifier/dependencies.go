package verifier

import (
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/presentproof/engine"
)

//go:generate mockery -inpkg -name=Provider
type Provider interface {
	Store() datastore.Store
	GetPresentationEngineRegistry() (engine.PresentationRegistry, error)
	GetPresentProofClient() (PresentProofClient, error)
}

//go:generate mockery -inpkg -name=PresentProofClient
type PresentProofClient interface {
	SendRequestPresentation(msg *ppclient.RequestPresentation, myDID, theirDID string) (string, error)
}
