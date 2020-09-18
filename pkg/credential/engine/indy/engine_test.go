package indy

import (
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/ursa"
)

func TestIssuerCredential(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := NewProvider()

		engine, err := New(prov)
		require.NoError(t, err)

		requestAttachment := decorator.AttachmentData{JSON: &ursa.CredentialOffer{}}

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
		requestAttachment := decorator.AttachmentData{JSON: &ursa.CredentialOffer{}}

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
		requestAttachment := decorator.AttachmentData{JSON: &ursa.CredentialOffer{}}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
	})

}
