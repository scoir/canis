package issuer

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	amqpmocks "github.com/scoir/canis/pkg/amqp/mocks"
	"github.com/scoir/canis/pkg/credential"
	emocks "github.com/scoir/canis/pkg/credential/engine/mocks"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
)

type handlerTestSuite struct {
	target                *CredHandler
	store                 *mocks.Store
	credsup               *credential.Supervisor
	registry              *emocks.CredentialRegistry
	notificationPublisher *amqpmocks.Publisher
}

type mockProp struct {
	myDID    string
	theirDID string
}

func (r *mockProp) MyDID() string {
	return r.myDID
}

func (r *mockProp) TheirDID() string {
	return r.theirDID
}

func (r *mockProp) All() map[string]interface{} {
	return map[string]interface{}{
		"MyDID":    r.myDID,
		"TheirDID": r.theirDID,
	}
}

func setup(t *testing.T) (*handlerTestSuite, func()) {
	suite := &handlerTestSuite{
		store:                 &mocks.Store{},
		registry:              &emocks.CredentialRegistry{},
		notificationPublisher: &amqpmocks.Publisher{},
	}

	suite.target = &CredHandler{
		store:                 suite.store,
		registry:              suite.registry,
		notificationPublisher: suite.notificationPublisher,
	}

	finish := func() {
		suite.store.AssertExpectations(t)
		suite.registry.AssertExpectations(t)
		suite.notificationPublisher.AssertExpectations(t)
	}

	return suite, finish
}

func TestCredHandler_ProposeCredentialMsg(t *testing.T) {
	t.Run("propose credential", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		schemaID := "schema-id"
		agent := &datastore.Agent{
			ID:                    "agent-id",
			EndorsableSchemaNames: []string{schemaID},
		}
		ac := &datastore.AgentConnection{}
		schema := &datastore.Schema{}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{
			Formats: []issuecredential.Format{
				{
					AttachID: "123",
					Format:   "hlindy-zkp-v1.0",
				},
			},
			FilterAttach: []decorator.Attachment{
				{
					ID:       "123",
					MimeType: "application/json",
					Data: decorator.AttachmentData{
						JSON: map[string]interface{}{},
					},
				},
			},
		}

		publishedMsg := `{"topic":"credentials","event":"proposed","message":{"agent_id":"agent-id","my_did":"","their_did":"","external_id":"","schema":{"ID":"","Format":"","Type":"","Name":"","Version":"","ExternalSchemaID":"","Context":null,"Attributes":null},"proposal":{}}}`
		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(agent, nil)
		suite.store.On("GetAgentConnectionForDID", agent, "did:their").Return(ac, nil)
		suite.registry.On("GetSchemaForProposal", "hlindy-zkp-v1.0", []byte(`{}`)).Return(schemaID, nil)
		suite.store.On("GetSchema", schemaID).Return(schema, nil)
		suite.store.On("InsertCredential", mock.AnythingOfType("*datastore.IssuedCredential")).Return("cred-id", nil)
		suite.notificationPublisher.On("Publish", []byte(publishedMsg), "application/json").Return(nil)

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential - error inserting", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		schemaID := "schema-id"
		agent := &datastore.Agent{
			ID:                    "agent-id",
			EndorsableSchemaNames: []string{schemaID},
		}
		ac := &datastore.AgentConnection{}
		schema := &datastore.Schema{}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{
			Formats: []issuecredential.Format{
				{
					AttachID: "123",
					Format:   "hlindy-zkp-v1.0",
				},
			},
			FilterAttach: []decorator.Attachment{
				{
					ID:       "123",
					MimeType: "application/json",
					Data: decorator.AttachmentData{
						JSON: map[string]interface{}{},
					},
				},
			},
		}

		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(agent, nil)
		suite.store.On("GetAgentConnectionForDID", agent, "did:their").Return(ac, nil)
		suite.registry.On("GetSchemaForProposal", "hlindy-zkp-v1.0", []byte(`{}`)).Return(schemaID, nil)
		suite.store.On("GetSchema", schemaID).Return(schema, nil)
		suite.store.On("InsertCredential", mock.AnythingOfType("*datastore.IssuedCredential")).Return("", errors.New("bad error"))

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential - can't find schema", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		schemaID := "schema-id"
		agent := &datastore.Agent{
			ID:                    "agent-id",
			EndorsableSchemaNames: []string{schemaID},
		}
		ac := &datastore.AgentConnection{}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{
			Formats: []issuecredential.Format{
				{
					AttachID: "123",
					Format:   "hlindy-zkp-v1.0",
				},
			},
			FilterAttach: []decorator.Attachment{
				{
					ID:       "123",
					MimeType: "application/json",
					Data: decorator.AttachmentData{
						JSON: map[string]interface{}{},
					},
				},
			},
		}

		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(agent, nil)
		suite.store.On("GetAgentConnectionForDID", agent, "did:their").Return(ac, nil)
		suite.registry.On("GetSchemaForProposal", "hlindy-zkp-v1.0", []byte(`{}`)).Return(schemaID, nil)
		suite.store.On("GetSchema", schemaID).Return(nil, errors.New("not found"))

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential - agent can't issue", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		schemaID := "schema-id"
		agent := &datastore.Agent{
			ID: "agent-id",
		}
		ac := &datastore.AgentConnection{}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{
			Formats: []issuecredential.Format{
				{
					AttachID: "123",
					Format:   "hlindy-zkp-v1.0",
				},
			},
			FilterAttach: []decorator.Attachment{
				{
					ID:       "123",
					MimeType: "application/json",
					Data: decorator.AttachmentData{
						JSON: map[string]interface{}{},
					},
				},
			},
		}
		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(agent, nil)
		suite.store.On("GetAgentConnectionForDID", agent, "did:their").Return(ac, nil)
		suite.registry.On("GetSchemaForProposal", "hlindy-zkp-v1.0", []byte(`{}`)).Return(schemaID, nil)

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential - no schema found", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		agent := &datastore.Agent{ID: "agent-id"}
		ac := &datastore.AgentConnection{}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{
			Formats:      []issuecredential.Format{},
			FilterAttach: []decorator.Attachment{},
		}
		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(agent, nil)
		suite.store.On("GetAgentConnectionForDID", agent, "did:their").Return(ac, nil)

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential bad their DID", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		agent := &datastore.Agent{ID: "agent-id"}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{
			Formats:      []issuecredential.Format{},
			FilterAttach: []decorator.Attachment{},
		}
		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(agent, nil)
		suite.store.On("GetAgentConnectionForDID", agent, "did:their").Return(nil, errors.New("not found"))

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential bad my DID", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
			Properties:   &mockProp{myDID: "did:my", theirDID: "did:their"},
		}
		proposal := &issuecredential.ProposeCredential{}
		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("not found"))
		suite.store.On("GetAgentByPublicDID", "did:my").Return(nil, errors.New("not found"))

		suite.target.ProposeCredentialMsg(action, proposal)

	})
	t.Run("propose credential negociation not supported", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		offer := &datastore.IssuedCredential{ProtocolID: thid}
		action := service.DIDCommAction{
			ProtocolName: "propose-credential",
			Message:      testMsg(t, thid),
			Stop:         func(error) {},
		}
		proposal := &issuecredential.ProposeCredential{}
		suite.store.On("FindCredentialByProtocolID", thid).Return(offer, nil)
		suite.store.On("DeleteCredentialByOffer", thid).Return(nil)

		suite.target.ProposeCredentialMsg(action, proposal)

	})
}

func TestRequestCredentialMsg(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
			Continue: func(args interface{}) {},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{
			AgentName:       "agent-01",
			SchemaName:      "schema-01",
			MyDID:           "did:sov:123",
			RegistryOfferID: "1234",
			Offer: &datastore.Offer{
				Preview: []issuecredential.Attribute{
					{Name: "attr1", Value: "test-val-1"},
				},
			},
		}
		agent := &datastore.Agent{}
		schema := &datastore.Schema{}
		did := &datastore.DID{}
		values := map[string]interface{}{
			"attr1": "test-val-1",
		}
		attach := &decorator.AttachmentData{JSON: map[string]interface{}{}}
		match := func(m *datastore.IssuedCredential) bool {
			return m.Credential != nil
		}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "agent-01").Return(agent, nil)
		suite.store.On("GetSchema", "schema-01").Return(schema, nil)
		suite.store.On("GetDID", "did:sov:123").Return(did, nil)
		suite.registry.On("IssueCredential", did, schema, "1234", request.RequestsAttach[0].Data, values).Return(attach, nil)
		suite.store.On("UpdateCredential", mock.MatchedBy(match)).Return(nil)

		suite.target.RequestCredentialMsg(action, request)
		require.NoError(t, err)
	})
	t.Run("unable to update credentual", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
			Continue: func(args interface{}) {},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{
			ID:              "cred-01",
			AgentName:       "agent-01",
			SchemaName:      "schema-01",
			MyDID:           "did:sov:123",
			RegistryOfferID: "1234",
			Offer: &datastore.Offer{
				Preview: []issuecredential.Attribute{
					{Name: "attr1", Value: "test-val-1"},
				},
			},
		}
		agent := &datastore.Agent{}
		schema := &datastore.Schema{}
		did := &datastore.DID{}
		values := map[string]interface{}{
			"attr1": "test-val-1",
		}
		attach := &decorator.AttachmentData{JSON: map[string]interface{}{}}
		match := func(m *datastore.IssuedCredential) bool {
			return m.Credential != nil
		}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "agent-01").Return(agent, nil)
		suite.store.On("GetSchema", "schema-01").Return(schema, nil)
		suite.store.On("GetDID", "did:sov:123").Return(did, nil)
		suite.registry.On("IssueCredential", did, schema, "1234", request.RequestsAttach[0].Data, values).Return(attach, nil)
		suite.store.On("UpdateCredential", mock.MatchedBy(match)).Return(errors.New("boom"))

		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "unexpected error updating issued credential cred-01: boom")
	})
	t.Run("empty result attachment data", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{
			AgentName:       "agent-01",
			SchemaName:      "schema-01",
			MyDID:           "did:sov:123",
			RegistryOfferID: "1234",
			Offer: &datastore.Offer{
				Preview: []issuecredential.Attribute{
					{Name: "attr1", Value: "test-val-1"},
				},
			},
		}
		agent := &datastore.Agent{}
		schema := &datastore.Schema{}
		did := &datastore.DID{}
		values := map[string]interface{}{
			"attr1": "test-val-1",
		}
		attach := &decorator.AttachmentData{}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "agent-01").Return(agent, nil)
		suite.store.On("GetSchema", "schema-01").Return(schema, nil)
		suite.store.On("GetDID", "did:sov:123").Return(did, nil)
		suite.registry.On("IssueCredential", did, schema, "1234", request.RequestsAttach[0].Data, values).Return(attach, nil)

		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "unable to fetch attachment: no contents in this attachment")
	})
	t.Run("registry error", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{
			AgentName:       "agent-01",
			SchemaName:      "schema-01",
			MyDID:           "did:sov:123",
			RegistryOfferID: "1234",
			Offer: &datastore.Offer{
				Preview: []issuecredential.Attribute{
					{Name: "attr1", Value: "test-val-1"},
				},
			},
		}
		agent := &datastore.Agent{}
		schema := &datastore.Schema{}
		did := &datastore.DID{}
		values := map[string]interface{}{
			"attr1": "test-val-1",
		}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "agent-01").Return(agent, nil)
		suite.store.On("GetSchema", "schema-01").Return(schema, nil)
		suite.store.On("GetDID", "did:sov:123").Return(did, nil)
		suite.registry.On("IssueCredential", did, schema, "1234", request.RequestsAttach[0].Data, values).Return(nil, errors.New("engine failure"))

		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "registry error creating credential: engine failure")
	})
	t.Run("bad DID", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{AgentName: "agent-01", SchemaName: "schema-01", MyDID: "bad"}
		agent := &datastore.Agent{}
		schema := &datastore.Schema{}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "agent-01").Return(agent, nil)
		suite.store.On("GetSchema", "schema-01").Return(schema, nil)
		suite.store.On("GetDID", "bad").Return(nil, errors.New("boom"))

		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "unable to find DID with ID bad: (boom)")
	})
	t.Run("bad schema", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{AgentName: "agent-01", SchemaName: "bad"}
		agent := &datastore.Agent{}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "agent-01").Return(agent, nil)
		suite.store.On("GetSchema", "bad").Return(nil, errors.New("not found"))

		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "unable to find schema with ID bad: (not found)")
	})
	t.Run("bad agent name", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"

		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		cred := &datastore.IssuedCredential{AgentName: "bad"}

		suite.store.On("FindCredentialByProtocolID", thid).Return(cred, nil)
		suite.store.On("GetAgent", "bad").Return(nil, errors.New("not found"))
		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "unable to find agent for credential request: not found")
	})
	t.Run("invalid credential protocol ID", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		thid := "80f8b418-4818-4af6-8915-f299b974f5c2"
		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Message:      testMsg(t, thid),
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{
			RequestsAttach: []decorator.Attachment{
				{Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				}},
			},
		}

		suite.store.On("FindCredentialByProtocolID", thid).Return(nil, errors.New("BOOM"))
		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "unable to find cred with ID 80f8b418-4818-4af6-8915-f299b974f5c2: (BOOM)")
	})
	t.Run("empty request", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "request-credential",
			Stop: func(e error) {
				err = e
			},
		}
		request := &issuecredential.RequestCredential{}
		suite.target.RequestCredentialMsg(action, request)
		require.Equal(t, err.Error(), "only one credential is supported at a time")
	})

}

func testMsg(t *testing.T, thid string) service.DIDCommMsg {
	msg, err := service.ParseDIDCommMsgMap([]byte(fmt.Sprintf(`{
						"@id":"80f8b418-4818-4af6-8915-f299b974f5c2",
						"@type":"https://didcomm.org/present-proof/2.0/request-presentation",
						"~thread":{
						   "thid":"%s"
						}
					}`, thid)))
	require.NoError(t, err)
	return msg
}
