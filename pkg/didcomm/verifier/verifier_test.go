package verifier

import (
	"context"
	"testing"

	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
	pemocks "github.com/scoir/canis/pkg/presentproof/engine/mocks"
	"github.com/scoir/canis/pkg/protogen/common"
	"github.com/scoir/canis/pkg/schema"
)

func TestServer_RequestPresentation(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Presentation: &common.RequestPresentation{
				Name:     "present-1",
				Version:  "0.0.0",
				SchemaId: "schema-1",
				RequestedAttributes: map[string]*common.AttrInfo{
					"attr1": {
						Name:         "attr1",
						Restrictions: "none",
					},
				},
				RequestedPredicates: map[string]*common.PredicateInfo{
					"pred1": {
						Name:         "pred1",
						PValue:       32,
						PType:        "int",
						Restrictions: "some",
					},
				},
			},
		}
		a := &datastore.Agent{Name: "agent-1"}
		ac := &datastore.AgentConnection{
			MyDID:    "sov:123",
			TheirDID: "sov:abc",
		}
		sch := &datastore.Schema{
			Name:   "schema-1",
			Format: "indy",
		}
		attrInfo := map[string]*schema.IndyProofRequestAttr{
			"attr1": {
				Name:         "attr1",
				Restrictions: "none",
			},
		}
		predicates := map[string]*schema.IndyProofRequestPredicate{
			"pred1": {
				Name:         "pred1",
				PValue:       32,
				PType:        "int",
				Restrictions: "some",
			},
		}
		data := &decorator.AttachmentData{JSON: map[string]interface{}{}}
		match := func(rp *ppclient.RequestPresentation) bool {
			return len(rp.Formats) == 1 && len(rp.RequestPresentationsAttach) == 1
		}
		prs := &datastore.PresentationRequest{
			AgentID:               "agent-1",
			SchemaID:              "schema-1",
			ExternalID:            req.ExternalId,
			PresentationRequestID: "ID-1",
			Data:                  []uint8(`{}`),
		}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-1").Return(sch, nil)
		suite.registry.On("RequestPresentation", "present-1", "0.0.0", "indy", attrInfo, predicates).Return(data, nil)
		suite.proofClient.On("SendRequestPresentation", mock.MatchedBy(match), "sov:123", "sov:abc").Return("ID-1", nil)
		suite.store.On("InsertPresentationRequest", prs).Return("ID-1", nil)

		res, err := suite.target.RequestPresentation(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "ID-1", res.RequestPresentationId)
	})
	t.Run("insert error", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Presentation: &common.RequestPresentation{
				Name:     "present-1",
				Version:  "0.0.0",
				SchemaId: "schema-1",
				RequestedAttributes: map[string]*common.AttrInfo{
					"attr1": {
						Name:         "attr1",
						Restrictions: "none",
					},
				},
				RequestedPredicates: map[string]*common.PredicateInfo{
					"pred1": {
						Name:         "pred1",
						PValue:       32,
						PType:        "int",
						Restrictions: "some",
					},
				},
			},
		}
		a := &datastore.Agent{Name: "agent-1"}
		ac := &datastore.AgentConnection{
			MyDID:    "sov:123",
			TheirDID: "sov:abc",
		}
		sch := &datastore.Schema{
			Name:   "schema-1",
			Format: "indy",
		}
		attrInfo := map[string]*schema.IndyProofRequestAttr{
			"attr1": {
				Name:         "attr1",
				Restrictions: "none",
			},
		}
		predicates := map[string]*schema.IndyProofRequestPredicate{
			"pred1": {
				Name:         "pred1",
				PValue:       32,
				PType:        "int",
				Restrictions: "some",
			},
		}
		data := &decorator.AttachmentData{JSON: map[string]interface{}{}}
		match := func(rp *ppclient.RequestPresentation) bool {
			return len(rp.Formats) == 1 && len(rp.RequestPresentationsAttach) == 1
		}
		prs := &datastore.PresentationRequest{
			AgentID:               "agent-1",
			SchemaID:              "schema-1",
			ExternalID:            req.ExternalId,
			PresentationRequestID: "ID-1",
			Data:                  []uint8(`{}`),
		}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-1").Return(sch, nil)
		suite.registry.On("RequestPresentation", "present-1", "0.0.0", "indy", attrInfo, predicates).Return(data, nil)
		suite.proofClient.On("SendRequestPresentation", mock.MatchedBy(match), "sov:123", "sov:abc").Return("ID-1", nil)
		suite.store.On("InsertPresentationRequest", prs).Return("", errors.New("unable to insert"))

		res, err := suite.target.RequestPresentation(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("client send error", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Presentation: &common.RequestPresentation{
				Name:     "present-1",
				Version:  "0.0.0",
				SchemaId: "schema-1",
				RequestedAttributes: map[string]*common.AttrInfo{
					"attr1": {
						Name:         "attr1",
						Restrictions: "none",
					},
				},
				RequestedPredicates: map[string]*common.PredicateInfo{
					"pred1": {
						Name:         "pred1",
						PValue:       32,
						PType:        "int",
						Restrictions: "some",
					},
				},
			},
		}
		a := &datastore.Agent{}
		ac := &datastore.AgentConnection{
			MyDID:    "sov:123",
			TheirDID: "sov:abc",
		}
		sch := &datastore.Schema{
			Format: "indy",
		}
		attrInfo := map[string]*schema.IndyProofRequestAttr{
			"attr1": {
				Name:         "attr1",
				Restrictions: "none",
			},
		}
		predicates := map[string]*schema.IndyProofRequestPredicate{
			"pred1": {
				Name:         "pred1",
				PValue:       32,
				PType:        "int",
				Restrictions: "some",
			},
		}
		data := &decorator.AttachmentData{}
		match := func(rp *ppclient.RequestPresentation) bool {
			return len(rp.Formats) == 1 && len(rp.RequestPresentationsAttach) == 1
		}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-1").Return(sch, nil)
		suite.registry.On("RequestPresentation", "present-1", "0.0.0", "indy", attrInfo, predicates).Return(data, nil)
		suite.proofClient.On("SendRequestPresentation", mock.MatchedBy(match), "sov:123", "sov:abc").Return("", errors.New("unable to send"))

		res, err := suite.target.RequestPresentation(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("engine failure", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Presentation: &common.RequestPresentation{
				Name:     "present-1",
				Version:  "0.0.0",
				SchemaId: "schema-1",
			},
		}
		a := &datastore.Agent{}
		ac := &datastore.AgentConnection{}
		sch := &datastore.Schema{
			Format: "indy",
		}
		attrInfo := map[string]*schema.IndyProofRequestAttr{}
		predicates := map[string]*schema.IndyProofRequestPredicate{}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-1").Return(sch, nil)
		suite.registry.On("RequestPresentation", "present-1", "0.0.0", "indy", attrInfo, predicates).Return(nil, errors.New("engine failure"))

		res, err := suite.target.RequestPresentation(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("bad schema", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
			Presentation: &common.RequestPresentation{
				SchemaId: "schema-1",
			},
		}
		a := &datastore.Agent{}
		ac := &datastore.AgentConnection{}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(ac, nil)
		suite.store.On("GetSchema", "schema-1").Return(nil, errors.New("not found"))

		res, err := suite.target.RequestPresentation(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("bad agent connection", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName:  "agent-1",
			ExternalId: "external-1",
		}
		a := &datastore.Agent{}

		suite.store.On("GetAgent", "agent-1").Return(a, nil)
		suite.store.On("GetAgentConnection", a, "external-1").Return(nil, errors.New("not found"))

		res, err := suite.target.RequestPresentation(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("bad agent", func(t *testing.T) {
		suite, cleanup := setupVerifier(t)
		defer cleanup()

		ctx := context.Background()
		req := &common.RequestPresentationRequest{
			AgentName: "agent-1",
		}

		suite.store.On("GetAgent", "agent-1").Return(nil, errors.New("not found"))

		res, err := suite.target.RequestPresentation(ctx, req)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

type verifierSuite struct {
	target      *Server
	store       *mocks.Store
	proofClient *MockPresentProofClient
	registry    *pemocks.PresentationRegistry
}

func setupVerifier(t *testing.T) (*verifierSuite, func()) {
	provider := &MockProvider{}

	out := &verifierSuite{
		store:       &mocks.Store{},
		proofClient: &MockPresentProofClient{},
		registry:    &pemocks.PresentationRegistry{},
	}

	provider.On("GetPresentProofClient").Return(out.proofClient, nil)
	provider.On("GetPresentationEngineRegistry").Return(out.registry, nil)
	provider.On("Store").Return(out.store)

	var err error
	out.target, err = New(provider)
	require.NoError(t, err)
	provider.AssertExpectations(t)

	return out, func() {
		out.proofClient.AssertExpectations(t)
		out.registry.AssertExpectations(t)
		out.store.AssertExpectations(t)
	}
}
