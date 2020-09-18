package indy

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/indy/mocks"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	storagemock "github.com/scoir/canis/pkg/mock/storage"
	"github.com/scoir/canis/pkg/ursa"
)

type providerMock struct {
	vdr    *mocks.IndyVDRClient
	store  *storagemock.MockStoreProvider
	issuer *issuermock
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
