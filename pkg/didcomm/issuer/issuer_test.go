package issuer

import (
	"context"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	_struct "github.com/golang/protobuf/ptypes/struct"

	"github.com/scoir/canis/pkg/credential/engine/mocks"
	"github.com/scoir/canis/pkg/datastore"
	dsmocks "github.com/scoir/canis/pkg/datastore/mocks"
	"github.com/scoir/canis/pkg/protogen/common"
)

func TestIssueCredential(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Credential: &common.Credential{
				Comment:  "test comment",
				SchemaId: "schema-2",
				Body:     &_struct.Struct{},
				Preview: []*common.CredentialAttribute{
					{
						Name:  "attr1",
						Value: "val1",
					},
				},
			},
		}
		a := &datastore.Agent{
			Name:      "agent-1",
			PublicDID: &datastore.DID{},
		}
		ac := &datastore.AgentConnection{
			MyDID:    "did:keri:abc",
			TheirDID: "did:keri:123",
		}
		sch := &datastore.Schema{}
		attach := &decorator.AttachmentData{}
		matcher := func(offer *issuecredential.OfferCredential) bool {
			return offer.Comment == "test comment"
		}
		cred := func(cred *datastore.IssuedCredential) bool {
			return cred.AgentName == "agent-1" &&
				cred.MyDID == "did:keri:abc" &&
				cred.TheirDID == "did:keri:123" &&
				cred.ProtocolID == "123"
		}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-2").Return(sch, nil)
		suite.registry.On("CreateCredentialOffer", a.PublicDID, "did:keri:123", sch, []byte(`{}`)).Return("1234", attach, nil)
		suite.issuer.On("SendOffer", mock.MatchedBy(matcher), "did:keri:abc", "did:keri:123").Return("123", nil)
		suite.store.On("InsertCredential", mock.MatchedBy(cred)).Return("abc", nil)

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "abc", res.CredentialId)
	})
	t.Run("insert cred failure", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Credential: &common.Credential{
				Comment:  "test comment",
				SchemaId: "schema-2",
				Body:     &_struct.Struct{},
				Preview: []*common.CredentialAttribute{
					{
						Name:  "attr1",
						Value: "val1",
					},
				},
			},
		}
		a := &datastore.Agent{
			PublicDID: &datastore.DID{},
		}
		ac := &datastore.AgentConnection{
			MyDID:    "did:keri:abc",
			TheirDID: "did:keri:123",
		}
		sch := &datastore.Schema{}
		attach := &decorator.AttachmentData{}
		matcher := func(offer *issuecredential.OfferCredential) bool {
			return offer.Comment == "test comment"
		}
		cred := func(cred *datastore.IssuedCredential) bool {
			return true
		}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-2").Return(sch, nil)
		suite.registry.On("CreateCredentialOffer", a.PublicDID, "did:keri:123", sch, []byte(`{}`)).Return("1234", attach, nil)
		suite.issuer.On("SendOffer", mock.MatchedBy(matcher), "did:keri:abc", "did:keri:123").Return("123", nil)
		suite.store.On("InsertCredential", mock.MatchedBy(cred)).Return("", errors.New("unable to save"))

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("send offer failure", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Credential: &common.Credential{
				Comment:  "test comment",
				SchemaId: "schema-2",
				Body:     &_struct.Struct{},
				Preview: []*common.CredentialAttribute{
					{
						Name:  "attr1",
						Value: "val1",
					},
				},
			},
		}
		a := &datastore.Agent{
			PublicDID: &datastore.DID{},
		}
		ac := &datastore.AgentConnection{
			MyDID:    "did:keri:abc",
			TheirDID: "did:keri:123",
		}
		sch := &datastore.Schema{}
		attach := &decorator.AttachmentData{}
		matcher := func(offer *issuecredential.OfferCredential) bool {
			return offer.Comment == "test comment"
		}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-2").Return(sch, nil)
		suite.registry.On("CreateCredentialOffer", a.PublicDID, "did:keri:123", sch, []byte(`{}`)).Return("1234", attach, nil)
		suite.issuer.On("SendOffer", mock.MatchedBy(matcher), "did:keri:abc", "did:keri:123").Return("", errors.New("boom"))

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("registry error", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Credential: &common.Credential{
				SchemaId: "schema-2",
				Body:     &_struct.Struct{},
			},
		}
		a := &datastore.Agent{
			PublicDID: &datastore.DID{},
		}
		ac := &datastore.AgentConnection{
			TheirDID: "did:keri:123",
		}
		sch := &datastore.Schema{}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-2").Return(sch, nil)
		suite.registry.On("CreateCredentialOffer", a.PublicDID, "did:keri:123", sch, []byte(`{}`)).Return("", nil, errors.New("registry failed"))

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("schema not found", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Credential: &common.Credential{
				SchemaId: "schema-2",
			},
		}
		a := &datastore.Agent{}
		ac := &datastore.AgentConnection{}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-2").Return(nil, errors.New("not found"))

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("agent connection not found", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
		}
		a := &datastore.Agent{}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(nil, errors.New("not found"))

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("agent not found", func(t *testing.T) {
		suite, cleanup := issuerSetup(t)
		defer cleanup()

		request := &common.IssueCredentialRequest{
			AgentName: "agent-1",
		}

		suite.store.On("GetAgent", "agent-1").Return(nil, errors.New("not found"))

		res, err := suite.target.IssueCredential(context.Background(), request)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

type suite struct {
	target   *Server
	store    *dsmocks.Store
	issuer   *MockCredentialIssuer
	registry *mocks.CredentialRegistry
}

func issuerSetup(t *testing.T) (*suite, func()) {
	ctx := &MockProvider{}
	out := &suite{
		store:    &dsmocks.Store{},
		issuer:   &MockCredentialIssuer{},
		registry: &mocks.CredentialRegistry{},
	}

	ctx.On("GetCredentialIssuer").Return(out.issuer, nil)
	ctx.On("GetCredentialEngineRegistry").Return(out.registry, nil)
	ctx.On("Store").Return(out.store)

	var err error
	out.target, err = New(ctx)
	require.NoError(t, err)

	ctx.AssertExpectations(t)

	return out, func() {
		out.issuer.AssertExpectations(t)
		out.registry.AssertExpectations(t)
	}
}
