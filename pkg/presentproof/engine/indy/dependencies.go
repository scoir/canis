package indy

import (
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	"github.com/scoir/canis/pkg/datastore"
)

//go:generate mockery -name=Oracle -inpkg
type Oracle interface {
	NewNonce() (string, error)
}

//go:generate mockery -name=Provider -inpkg
type Provider interface {
	IndyVDR() (VDRClient, error)
	Store() datastore.Store
	Oracle() Oracle
}

//go:generate mockery -name=VDRClient -inpkg
type VDRClient interface {
	GetCredDef(credDefID string) (*vdr.ReadReply, error)
}
