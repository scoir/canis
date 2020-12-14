package indy

import (
	"fmt"
	"testing"

	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	kmsMock "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
)

func TestIssuerCredential(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		requestAttachment := decorator.AttachmentData{JSON: &schema.IndyCredentialOffer{}}

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
		prov.store.Store.Store["test-offer-id"] = []byte(`{"cred_def_id": "test-creddef-id"}`)
		prov.store.Store.Store["test-creddef-id"] = []byte("{}")
		var values map[string]interface{}

		prov.vdr.On("GetCredDef", "test-creddef-id").Return(&vdr.ReadReply{Data: map[string]interface{}{
			"primary": map[string]interface{}{},
		}}, nil)
		prov.issuer.IssueCredentialAttachment = &decorator.AttachmentData{}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.NoError(t, err)
		require.Equal(t, prov.issuer.IssueCredentialAttachment, attachment)
	})
	t.Run("invalid attachment", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{}
		offerID := "test-offer-id"
		var values map[string]interface{}
		s := &datastore.Schema{}
		requestAttachment := decorator.AttachmentData{}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
		fmt.Println(err)
	})

	t.Run("bad offer ID", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{}
		offerID := "test-offer-id"
		var values map[string]interface{}
		s := &datastore.Schema{}
		requestAttachment := decorator.AttachmentData{JSON: &schema.IndyCredentialOffer{}}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
	})

	t.Run("bad cred def ID", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		issuerDID := &datastore.DID{}
		offerID := "test-offer-id"
		prov.store.Store.Store["test-offer-id"] = []byte(`{"cred_def_id": "test-creddef-id"}`)
		prov.vdr.On("GetCredDef", "test-creddef-id").Return(&vdr.ReadReply{Data: map[string]interface{}{
			"primary": map[string]interface{}{},
		}}, nil)

		var values map[string]interface{}
		s := &datastore.Schema{}
		requestAttachment := decorator.AttachmentData{JSON: &schema.IndyCredentialOffer{}}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
	})

}

func TestAccept(t *testing.T) {
	prov := NewProvider()

	engine, err := New(prov)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		ac := engine.Accept("hlindy-zkp-v1.0")
		require.True(t, ac)
	})
	t.Run("sad", func(t *testing.T) {
		ac := engine.Accept("lds/ld-proof")
		require.False(t, ac)
	})
}

func TestCreateSchema(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)
		issuer := &datastore.DID{
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
		s := &datastore.Schema{
			Name:    "schema-name",
			Version: "1.2",
			Attributes: []*datastore.Attribute{
				{
					Name: "field1",
					Type: 0,
				},
				{
					Name: "field2",
					Type: 0,
				},
			},
		}

		kh, err := kmsMock.CreateMockED25519KeyHandle()
		require.NoError(t, err)
		prim, err := kh.Primitives()
		require.NoError(t, err)
		mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)

		prov.kms.GetKeyValue = kh

		prov.vdr.On("CreateSchema", issuer.DID.MethodID(), "schema-name", "1.2", []string{"field1", "field2"}, mysig).Return("test-schema-id", nil)

		sid, err := engine.CreateSchema(issuer, s)
		require.NoError(t, err)
		require.Equal(t, "test-schema-id", sid)

	})
	t.Run("no keypair found", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)
		issuer := &datastore.DID{
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

		prov.kms.GetKeyErr = errors.New("not found")

		sid, err := engine.CreateSchema(issuer, s)
		require.Error(t, err)
		require.Empty(t, sid)

	})
	t.Run("vdr failure", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)
		issuer := &datastore.DID{
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
		s := &datastore.Schema{
			Name:    "schema-name",
			Version: "1.2",
			Attributes: []*datastore.Attribute{
				{
					Name: "field1",
					Type: 0,
				},
				{
					Name: "field2",
					Type: 0,
				},
			},
		}

		kh, err := kmsMock.CreateMockED25519KeyHandle()
		require.NoError(t, err)
		prim, err := kh.Primitives()
		require.NoError(t, err)
		mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)

		prov.kms.GetKeyValue = kh

		prov.vdr.On("CreateSchema", issuer.DID.MethodID(), "schema-name", "1.2", []string{"field1", "field2"}, mysig).Return("", errors.New("BOOM"))

		sid, err := engine.CreateSchema(issuer, s)
		require.Error(t, err)
		require.Empty(t, sid)

	})
}

func TestGetSchemaForProposal(t *testing.T) {
	t.Run("get schema", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		proposal := []byte(`{"schema_id": "123"}`)
		schemaID, err := engine.GetSchemaForProposal(proposal)
		require.NoError(t, err)
		require.Equal(t, "123", schemaID)
	})
	t.Run("get schema - bad JSON", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		proposal := []byte(`{"schema_id": "`)
		schemaID, err := engine.GetSchemaForProposal(proposal)
		require.Error(t, err)
		require.Equal(t, "", schemaID)
	})
}
