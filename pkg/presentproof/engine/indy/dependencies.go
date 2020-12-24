package indy

import (
	"github.com/scoir/canis/pkg/credential/engine/indy"
	"github.com/scoir/canis/pkg/datastore"
)

//go:generate mockery -name=Provider -inpkg
type Provider interface {
	IndyVDR() (indy.VDRClient, error)
	Store() datastore.Store
	Oracle() indy.Oracle
}
