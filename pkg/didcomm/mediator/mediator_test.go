package mediator

import (
	"context"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
	medmocks "github.com/scoir/canis/pkg/didcomm/mediator/mocks"
	"github.com/scoir/canis/pkg/protogen/common"
)

type MediatorTestSuite struct {
	Store *mocks.Store
}

func SetupTest(t *testing.T) (*Mediator, *MediatorTestSuite) {
	suite := &MediatorTestSuite{}
	suite.Store = &mocks.Store{}

	ar, err := aries.New(aries.WithStoreProvider(mem.NewProvider()))
	require.NoError(t, err)

	prov := &medmocks.Provider{}

	prov.On("GetExternal").Return("test-external")
	prov.On("GetEdgeAgentSecret").Return("secret words")
	prov.On("GetDatastore").Return(suite.Store, nil)
	prov.On("GetAriesContext").Return(ar.Context())

	target, err := New(prov)
	require.NoError(t, err)
	require.NotNil(t, target)
	prov.AssertExpectations(t)

	return target, suite
}

func TestNew(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		ar, err := aries.New(aries.WithStoreProvider(mem.NewProvider()))
		require.NoError(t, err)

		prov := &medmocks.Provider{}

		prov.On("GetExternal").Return("test-external")
		prov.On("GetEdgeAgentSecret").Return("test-secret")
		prov.On("GetDatastore").Return(&mocks.Store{}, nil)
		prov.On("GetAriesContext").Return(ar.Context())

		target, err := New(prov)
		require.NoError(t, err)
		require.NotNil(t, target)

		prov.AssertExpectations(t)

	})

	t.Run("database error", func(t *testing.T) {
		prov := &medmocks.Provider{}

		prov.On("GetExternal").Return("test-external")
		prov.On("GetEdgeAgentSecret").Return("test-secret")
		prov.On("GetDatastore").Return(nil, errors.New("boom"))

		target, err := New(prov)
		require.Error(t, err)
		require.Nil(t, target)

		prov.AssertExpectations(t)
	})

	t.Run("aries error", func(t *testing.T) {
		prov := &medmocks.Provider{}

		prov.On("GetExternal").Return("test-external")
		prov.On("GetEdgeAgentSecret").Return("test-secret")
		prov.On("GetDatastore").Return(&mocks.Store{}, nil)
		prov.On("GetAriesContext").Return(nil, errors.New("boom"))

		target, err := New(prov)
		require.Error(t, err)
		require.Nil(t, target)

		prov.AssertExpectations(t)
	})
}

func TestRegisterEdgeAgent(t *testing.T) {
	t.Run("bad secret", func(t *testing.T) {
		target, _ := SetupTest(t)

		req := &common.RegisterEdgeAgentRequest{Secret: "not correct"}

		resp, err := target.RegisterEdgeAgent(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, "rpc error: code = Unauthenticated desc = invalid edge agent secret", err.Error())
		require.Nil(t, resp)

	})

	t.Run("not seeded", func(t *testing.T) {
		target, suite := SetupTest(t)

		suite.Store.On("GetMediatorDID").Return(nil, errors.New("not found"))

		req := &common.RegisterEdgeAgentRequest{Secret: "secret words"}

		resp, err := target.RegisterEdgeAgent(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, "rpc error: code = Internal desc = unable to load mediator Public with DID, system must be seeded with an identity", err.Error())
		require.Nil(t, resp)

		suite.Store.AssertExpectations(t)
	})

	t.Run("failed datastore registration", func(t *testing.T) {
		target, suite := SetupTest(t)

		mediatorDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "abc123",
					Method:           "sov",
				},
				Verkey: "abc",
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: base58.Encode([]byte{1, 2, 3}),
			},
			Endpoint: "ws://test:1001",
		}

		connIDMatch := func(val string) bool {
			return len(val) > 0
		}

		suite.Store.On("GetMediatorDID").Return(mediatorDID, nil)
		suite.Store.On("RegisterEdgeAgent", mock.MatchedBy(connIDMatch), "ext-test").Return("", errors.New("boom"))

		req := &common.RegisterEdgeAgentRequest{Secret: "secret words", ExternalId: "ext-test"}

		resp, err := target.RegisterEdgeAgent(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)

		suite.Store.AssertExpectations(t)
	})

	t.Run("happy path", func(t *testing.T) {
		target, suite := SetupTest(t)

		mediatorDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "abc123",
					Method:           "sov",
				},
				Verkey: "abc",
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: base58.Encode([]byte{1, 2, 3}),
			},
			Endpoint: "ws://test:1001",
		}

		connIDMatch := func(val string) bool {
			return len(val) > 0
		}

		suite.Store.On("GetMediatorDID").Return(mediatorDID, nil)
		suite.Store.On("RegisterEdgeAgent", mock.MatchedBy(connIDMatch), "ext-test").Return("success-id", nil)

		req := &common.RegisterEdgeAgentRequest{Secret: "secret words", ExternalId: "ext-test"}

		resp, err := target.RegisterEdgeAgent(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, "success-id", resp.Id)

		suite.Store.AssertExpectations(t)
	})
}

func TestEndpoint(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		target, _ := SetupTest(t)

		req := &common.EndpointRequest{}
		endpoint, err := target.GetEndpoint(context.Background(), req)
		require.NoError(t, err)

		require.Equal(t, "test-external", endpoint.Endpoint)
	})
}

func TestAccept(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		target, suite := SetupTest(t)

		f := target.accepted("test-external")

		conn := &didexchange.Connection{
			Record: &connection.Record{
				ConnectionID: "conn-id",
				TheirDID:     "did:peer:xyz",
				MyDID:        "did:sov:abc",
			},
		}

		ea := &datastore.EdgeAgent{
			ExternalID: "test-external",
		}
		updated := &datastore.EdgeAgent{
			TheirDID:   "did:peer:xyz",
			MyDID:      "did:sov:abc",
			ExternalID: "test-external",
		}
		suite.Store.On("GetEdgeAgent", "conn-id").Return(ea, nil)
		suite.Store.On("UpdateEdgeAgent", updated).Return(nil)

		f("id", conn)

		suite.Store.AssertExpectations(t)

	})

	t.Run("failed get", func(t *testing.T) {
		target, suite := SetupTest(t)

		f := target.accepted("test-external")

		conn := &didexchange.Connection{
			Record: &connection.Record{
				ConnectionID: "conn-id",
				TheirDID:     "did:peer:xyz",
				MyDID:        "did:sov:abc",
			},
		}

		suite.Store.On("GetEdgeAgent", "conn-id").Return(nil, errors.New("boom"))

		f("id", conn)
		suite.Store.AssertExpectations(t)

	})

	t.Run("wrong external ID", func(t *testing.T) {
		target, suite := SetupTest(t)

		f := target.accepted("test-external")

		conn := &didexchange.Connection{
			Record: &connection.Record{
				ConnectionID: "conn-id",
				TheirDID:     "did:peer:xyz",
				MyDID:        "did:sov:abc",
			},
		}

		ea := &datastore.EdgeAgent{
			ExternalID: "bad",
		}
		suite.Store.On("GetEdgeAgent", "conn-id").Return(ea, nil)

		f("id", conn)
		suite.Store.AssertExpectations(t)

	})

	t.Run("failed update", func(t *testing.T) {
		target, suite := SetupTest(t)

		f := target.accepted("test-external")

		conn := &didexchange.Connection{
			Record: &connection.Record{
				ConnectionID: "conn-id",
				TheirDID:     "did:peer:xyz",
				MyDID:        "did:sov:abc",
			},
		}

		ea := &datastore.EdgeAgent{
			ExternalID: "test-external",
		}
		updated := &datastore.EdgeAgent{
			TheirDID:   "did:peer:xyz",
			MyDID:      "did:sov:abc",
			ExternalID: "test-external",
		}
		suite.Store.On("GetEdgeAgent", "conn-id").Return(ea, nil)
		suite.Store.On("UpdateEdgeAgent", updated).Return(errors.New("boom"))

		f("id", conn)
		suite.Store.AssertExpectations(t)

	})

}

func TestFailed(t *testing.T) {
	failed("id", errors.New("boom"))
}
