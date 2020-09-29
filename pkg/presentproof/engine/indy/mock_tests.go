package indy

import (
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	storagemock "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/indy/mocks"
	"github.com/scoir/canis/pkg/ursa"
)

type providerMock struct {
	vdr      *mocks.IndyVDRClient
	store    *storagemock.MockStoreProvider
	verifier *verifiermock
}

func NewProvider() *providerMock {
	return &providerMock{
		vdr: &mocks.IndyVDRClient{},
		store: &storagemock.MockStoreProvider{
			Store: &storagemock.MockStore{
				Store: map[string][]byte{},
			},
		},
		verifier: &verifiermock{},
	}
}

func (r *providerMock) IndyVDR() (indy.IndyVDRClient, error) {
	return r.vdr, nil
}

func (r *providerMock) KMS() (kms.KeyManager, error) {
	return nil, nil
}

func (r *providerMock) StorageProvider() storage.Provider {
	return r.store
}

func (r *providerMock) Verifier() ursa.Verifier {
	return r.verifier
}

type verifiermock struct {
}

func (v verifiermock) SendRequestPresentation(msg *ppprotocol.RequestPresentation, myDID, theirDID string) (string, error) {
	panic("implement me")
}
