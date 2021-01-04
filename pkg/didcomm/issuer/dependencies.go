package issuer

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"

	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
)

//go:generate mockery -inpkg -name=Provider
type Provider interface {
	Store() datastore.Store
	GetCredentialIssuer() (CredentialIssuer, error)
	GetCredentialEngineRegistry() (engine.CredentialRegistry, error)
}

//go:generate mockery -inpkg -name=CredentialIssuer
type CredentialIssuer interface {
	SendOffer(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error)
}
