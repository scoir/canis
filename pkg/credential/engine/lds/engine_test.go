package lds

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	kmsMock "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	storeMock "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	vdriMock "github.com/hyperledger/aries-framework-go/pkg/mock/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/scoir/canis/pkg/datastore"
)

type mockProvider struct {
	kms      *kmsMock.KeyManager
	store    *storeMock.MockStoreProvider
	registry *vdriMock.MockVDRIRegistry
}

func newProv() *mockProvider {
	return &mockProvider{
		kms:      &kmsMock.KeyManager{},
		store:    storeMock.NewMockStoreProvider(),
		registry: &vdriMock.MockVDRIRegistry{},
	}
}

func (r *mockProvider) KMS() kms.KeyManager {
	return r.kms
}

func (r *mockProvider) StorageProvider() storage.Provider {
	return r.store
}

func (r *mockProvider) VDRIRegistry() vdriapi.Registry {
	return r.registry
}

func TestIssueCredential(t *testing.T) {
	vals := map[string]interface{}{
		"firstName": "Bilbo",
		"lastName":  "Baggins",
	}
	req := decorator.AttachmentData{JSON: vals}
	offer := &credOffer{
		Values: vals,
	}
	d, _ := json.Marshal(offer)

	t.Run("issue correctly", func(t *testing.T) {
		prov := newProv()
		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: "test",
			},
		}
		s := &datastore.Schema{}
		offerID := "test-offer-id"
		doc := &diddoc.Doc{PublicKey: []diddoc.PublicKey{
			{
				ID: "#",
			},
		}}
		kh, err := kmsMock.CreateMockED25519KeyHandle()
		require.NoError(t, err)

		prov.store.Store.Store["test-offer-id"] = d
		prov.registry.ResolveValue = doc
		prov.kms.GetKeyValue = kh

		res, err := engine.IssueCredential(issuerDID, s, offerID, req, vals)
		require.NoError(t, err)
		require.NotNil(t, res)

	})
	t.Run("unable to resolve DID", func(t *testing.T) {
		prov := newProv()
		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}
		s := &datastore.Schema{}
		offerID := "test-offer-id"
		prov.store.Store.Store["test-offer-id"] = d

		res, err := engine.IssueCredential(issuerDID, s, offerID, req, vals)
		require.Nil(t, res)
		require.Error(t, err)

	})
	t.Run("request values don't match offer values", func(t *testing.T) {
		prov := newProv()
		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}
		s := &datastore.Schema{}
		offerID := "test-offer-id"
		prov.store.Store.Store["test-offer-id"] = []byte(`{}`)

		res, err := engine.IssueCredential(issuerDID, s, offerID, req, vals)
		require.Nil(t, res)
		require.Error(t, err)

	})
	t.Run("invalid offer", func(t *testing.T) {
		prov := newProv()
		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}
		s := &datastore.Schema{}
		offerID := "test-offer-id"

		res, err := engine.IssueCredential(issuerDID, s, offerID, req, vals)
		require.Nil(t, res)
		require.Error(t, err)

	})
}

func TestOfferCredential(t *testing.T) {
	vals := `{
		"firstName": "Bilbo",
		"lastName":  "Baggins"
	}`

	t.Run("offer credential", func(t *testing.T) {
		prov := newProv()
		engine, err := New(prov)
		require.NoError(t, err)

		subjectDID := "did:scr:S1uRyT6S3GyYCC4Q4ryirH"
		s := &datastore.Schema{}

		offerID, attach, err := engine.CreateCredentialOffer(nil, subjectDID, s, []byte(vals))
		require.NoError(t, err)
		require.NotEmpty(t, offerID)
		require.Equal(t, "eyJAY29udGV4dCI6bnVsbCwiQHR5cGUiOlsiIl0sImZpcnN0TmFtZSI6IkJpbGJvIiwibGFzdE5hbWUiOiJCYWdnaW5zIn0=", attach.Base64)
	})
}
