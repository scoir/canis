package indy

import (
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	storagemock "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/indy/mocks"
	"github.com/scoir/canis/pkg/ursa"
)

type providerMock struct {
	vdr            *mocks.IndyVDRClient
	vdrError       error
	kms            kms.KeyManager
	store          *storeMock
	cryptoProvider *cryptoMock
}

// IndyVDR mock implementation for indy engine
func (r *providerMock) IndyVDR() (indy.IndyVDRClient, error) {
	if r.vdrError != nil {
		return nil, r.vdrError
	}

	return r.vdr, nil
}

// KMS mock implementation for indy engine
func (r *providerMock) KMS() kms.KeyManager {
	return r.kms
}

// StorageProvider mock implementation for indy engine
func (r *providerMock) StorageProvider() storage.Provider {
	return r.store
}

// Verifier mock implementation for indy engine
func (r *providerMock) Verifier() ursa.Verifier {
	return nil
}

type storeMock struct {
	OpenStoreErr error
	Store        storage.Store
}

// OpenStore mock implementation for store mock
func (r *storeMock) OpenStore(name string) (storage.Store, error) {
	if r.OpenStoreErr != nil {
		return nil, r.OpenStoreErr
	}

	//todo: revisit using aries mocks here
	return &storagemock.MockStore{}, nil
}

// CloseStore unimplemented for store mock
func (r *storeMock) CloseStore(name string) error {
	return nil
}

// Close unimplemented for store mock
func (r *storeMock) Close() error {
	return nil
}

type cryptoMock struct {
	NewNonceErr error
}

// NewNonce mock implementation for crypto
func (r *cryptoMock) NewNonce() (string, error) {
	if r.NewNonceErr != nil {
		return "", r.NewNonceErr
	}

	return "nonce", nil
}
