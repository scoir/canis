package indy

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	kmsMock "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	storagemock "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/indy/mocks"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/ursa"
)

type providerMock struct {
	vdr    *mocks.IndyVDRClient
	store  *storagemock.MockStoreProvider
	issuer *issuermock
	kms    *kmsMock.KeyManager
}

func NewProvider() *providerMock {
	return &providerMock{
		vdr: &mocks.IndyVDRClient{},
		store: &storagemock.MockStoreProvider{
			Store: &storagemock.MockStore{
				Store: map[string][]byte{},
			},
		},
		issuer: &issuermock{},
		kms:    &kmsMock.KeyManager{},
	}
}

func (r *providerMock) IndyVDR() (indy.IndyVDRClient, error) {
	return r.vdr, nil
}

func (r *providerMock) KMS() kms.KeyManager {
	return r.kms
}

func (r *providerMock) StorageProvider() storage.Provider {
	return r.store
}

func (r *providerMock) Issuer() ursa.Issuer {
	return r.issuer
}

type issuermock struct {
	IssueCredentialAttachment *decorator.AttachmentData
	IssueCredentialError      error
}

func (r *issuermock) IssueCredential(issuerDID string, schemaID, credDefID, offerNonce string, blindedMasterSecret, blindedMSCorrectnessProof, requestNonce string, credDef *vdr.ClaimDefData, credDefPrivateKey string, values map[string]interface{}) (*decorator.AttachmentData, error) {
	return r.IssueCredentialAttachment, r.IssueCredentialError
}
