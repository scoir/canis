package indy

import "C"
import (
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/ursa"
)

const Indy = "indy"

type PresentationEngine struct {
	client   VDRClient
	kms      kms.KeyManager
	store    store
	verifier UrsaVerifier
}

type provider interface {
	IndyVDR() (indy.IndyVDRClient, error)
	KMS() (kms.KeyManager, error)
	StorageProvider() storage.Provider
	Verifier() ursa.Verifier
}

type UrsaVerifier interface {
	SendRequestPresentation(msg *ppprotocol.RequestPresentation, myDID, theirDID string) (string, error)
}

type store interface {
	Get(k string) ([]byte, error)
	Put(k string, v []byte) error
}

type VDRClient interface {
}

func New(prov provider) (*PresentationEngine, error) {
	eng := &PresentationEngine{}

	var err error
	eng.store, err = prov.StorageProvider().OpenStore("indy_engine")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store for indy engine")
	}

	eng.client, err = prov.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr for indy proof engine")
	}

	eng.kms, err = prov.KMS()
	if err != nil {
		return nil, err
	}

	eng.verifier = prov.Verifier()

	return eng, nil
}

func (r *PresentationEngine) Accept(typ string) bool {
	return typ == Indy
}
