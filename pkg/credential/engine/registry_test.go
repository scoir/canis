package engine

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/stretchr/testify/require"

	emocks "github.com/scoir/canis/pkg/credential/engine/mocks"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
)

type providerMock struct {
	store *mocks.Store
}

func (r *providerMock) Store() datastore.Store {
	return r.store
}

func NewProvider() *providerMock {
	return &providerMock{store: &mocks.Store{}}
}

func TestCreateCredentialOffer(t *testing.T) {
	prov := NewProvider()
	eng := &emocks.CredentialEngine{
		Accep:                           true,
		CredentialOfferID:               "test-offer-id",
		CreateCredentialOfferAttachment: &decorator.AttachmentData{},
		CreateCredentialOfferError:      nil,
	}
	reg := New(prov, WithEngine(eng))

	did := &datastore.DID{}
	s := &datastore.Schema{Type: "indy"}
	id, attach, err := reg.CreateCredentialOffer(did, s)
	require.Equal(t, id, eng.CredentialOfferID)
	require.Equal(t, attach, eng.CreateCredentialOfferAttachment)
	require.NoError(t, err)

	eng.CredentialOfferID = ""
	eng.CreateCredentialOfferAttachment = nil
	eng.CreateCredentialOfferError = errors.New("BOOM")
	id, attach, err = reg.CreateCredentialOffer(did, s)
	require.Empty(t, id)
	require.Nil(t, attach)
	require.Error(t, err)
	require.Equal(t, err.Error(), "BOOM")

}

func TestIssueCredential(t *testing.T) {
	prov := NewProvider()
	eng := &emocks.CredentialEngine{
		Accep:                     true,
		IssueCredentialAttachment: &decorator.AttachmentData{},
		IssueCredentialError:      nil,
	}
	reg := New(prov, WithEngine(eng))

	did := &datastore.DID{}
	s := &datastore.Schema{Type: "indy"}
	attach, err := reg.IssueCredential(did, s, "test", decorator.AttachmentData{}, map[string]interface{}{})
	require.Equal(t, attach, eng.IssueCredentialAttachment)
	require.NoError(t, err)

	eng.IssueCredentialAttachment = nil
	eng.IssueCredentialError = errors.New("BOOM")
	attach, err = reg.IssueCredential(did, s, "test", decorator.AttachmentData{}, map[string]interface{}{})
	require.Nil(t, attach)
	require.Error(t, err)
	require.Equal(t, err.Error(), "BOOM")

}

func TestRegisterSchema(t *testing.T) {
	prov := NewProvider()
	eng := &emocks.CredentialEngine{
		Accep: true,
	}
	reg := New(prov, WithEngine(eng))

	did := &datastore.DID{}
	s := &datastore.Schema{Type: "indy"}
	err := reg.RegisterSchema(did, s)
	require.NoError(t, err)

	eng.RegisterError = errors.New("BOOM")
	err = reg.RegisterSchema(did, s)
	require.Error(t, err)
	require.Equal(t, err.Error(), "error from credential engine: BOOM")

}

func TestCreateSchema(t *testing.T) {
	prov := NewProvider()
	eng := &emocks.CredentialEngine{
		Accep:             true,
		SchemaID:          "test-schema-id",
		CreateSchemaError: nil,
	}
	reg := New(prov, WithEngine(eng))

	did := &datastore.DID{}
	prov.store.On("GetPublicDID").Return(did, nil).Once()
	s := &datastore.Schema{Type: "indy"}
	id, err := reg.CreateSchema(s)
	require.Equal(t, id, eng.SchemaID)
	require.NoError(t, err)

	prov.store.On("GetPublicDID").Return(did, nil).Once()
	eng.SchemaID = ""
	eng.CreateSchemaError = errors.New("BOOM")
	id, err = reg.CreateSchema(s)
	require.Empty(t, id)
	require.Error(t, err)
	require.Equal(t, err.Error(), "error from credential engine: BOOM")

	prov.store.On("GetPublicDID").Return(nil, errors.New("NO DID")).Once()
	id, err = reg.CreateSchema(s)
	require.Empty(t, id)
	require.Error(t, err)
	require.Equal(t, err.Error(), "error getting public did to create schema: NO DID")

}

func TestNoValidEngine(t *testing.T) {
	prov := NewProvider()
	eng := &emocks.CredentialEngine{Accep: false}
	reg := New(prov, WithEngine(eng))

	did := &datastore.DID{}
	s := &datastore.Schema{Type: "indy"}
	id, attach, err := reg.CreateCredentialOffer(did, s)
	require.Empty(t, id)
	require.Nil(t, attach)
	require.Error(t, err)
	require.Equal(t, err.Error(), "credential type indy not supported by any engine")

	attach, err = reg.IssueCredential(did, s, "test", decorator.AttachmentData{}, map[string]interface{}{})
	require.Nil(t, attach)
	require.Error(t, err)
	require.Equal(t, err.Error(), "credential type indy not supported by any engine")

	err = reg.RegisterSchema(did, s)
	require.Error(t, err)
	require.Equal(t, err.Error(), "credential type indy not supported by any engine")

	id, err = reg.CreateSchema(s)
	require.Empty(t, id)
	require.Error(t, err)
	require.Equal(t, err.Error(), "credential type indy not supported by any engine")

}
