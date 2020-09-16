/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/apiserver/api"
	apimocks "github.com/scoir/canis/pkg/apiserver/mocks"
	emocks "github.com/scoir/canis/pkg/credential/engine/mocks"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
	lb "github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
	dmocks "github.com/scoir/canis/pkg/didexchange/mocks"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
)

type AdminTestSuite struct {
	Store             *mocks.Store
	Bouncer           *dmocks.Bouncer
	CredRegistry      *emocks.CredentialRegistry
	LoadbalanceClient lb.LoadbalancerClient
	IndyClient        *apimocks.MockVDRClient
}

func SetupTest() (*APIServer, *AdminTestSuite) {
	suite := &AdminTestSuite{}
	suite.Store = &mocks.Store{}
	suite.Bouncer = &dmocks.Bouncer{}
	suite.CredRegistry = &emocks.CredentialRegistry{}
	suite.IndyClient = &apimocks.MockVDRClient{}

	target := &APIServer{
		agentStore:     suite.Store,
		schemaStore:    suite.Store,
		didStore:       suite.Store,
		client:         suite.IndyClient,
		schemaRegistry: suite.CredRegistry,
	}

	return target, suite
}

func TestCreateAgent(t *testing.T) {

	t.Run("no public did", func(t *testing.T) {
		target, suite := SetupTest()
		request := &api.CreateAgentRequest{
			Agent: &api.Agent{
				Id:                  "123",
				Name:                "Test Agent",
				AssignedSchemaId:    "",
				EndorsableSchemaIds: nil,
			},
		}

		a := &datastore.Agent{
			ID:                  "123",
			Name:                "Test Agent",
			AssignedSchemaId:    "",
			EndorsableSchemaIds: nil,
		}

		suite.Store.On("GetAgent", "123").Return(nil, errors.New("not found"))
		suite.Store.On("InsertAgent", a).Return("123", nil)

		resp, err := target.CreateAgent(context.Background(), request)
		require.Nil(t, err)
		require.Equal(t, resp.Id, "123")
	})
	t.Run("with public did", func(t *testing.T) {
		target, suite := SetupTest()
		mocklb := &apimocks.MockLoadbalancer{
			EndpointValue: "0.0.0.0:420",
		}
		target.loadbalancer = mocklb

		request := &api.CreateAgentRequest{
			Agent: &api.Agent{
				Id:                  "123",
				Name:                "Test Agent",
				AssignedSchemaId:    "",
				EndorsableSchemaIds: nil,
				PublicDid:           true,
			},
		}

		d, keypair, err := identifiers.CreateDID(&identifiers.MyDIDInfo{
			Cid:        true,
			MethodName: "scr",
		})
		require.NoError(t, err)

		did := &datastore.DID{
			DID: d,
			KeyPair: &datastore.KeyPair{
				PublicKey:  keypair.PublicKey(),
				PrivateKey: keypair.PrivateKey(),
			},
		}

		suite.Store.On("GetAgent", "123").Return(nil, errors.New("not found"))
		suite.Store.On("InsertAgent", mock.AnythingOfType("*datastore.Agent")).Return("123", nil)
		suite.Store.On("GetPublicDID").Return(did, nil)

		resp, err := target.CreateAgent(context.Background(), request)
		assert.Nil(t, err)
		assert.Equal(t, resp.Id, "123")
	})
}

func TestCreateAgentFails(t *testing.T) {
	target, suite := SetupTest()
	request := &api.CreateAgentRequest{
		Agent: &api.Agent{
			Id:                  "123",
			Name:                "Test Agent",
			AssignedSchemaId:    "",
			EndorsableSchemaIds: nil,
		},
	}

	a := &datastore.Agent{
		ID:                  "123",
		Name:                "Test Agent",
		AssignedSchemaId:    "",
		EndorsableSchemaIds: nil,
	}

	suite.Store.On("GetAgent", "123").Return(nil, errors.New("not found"))
	suite.Store.On("InsertAgent", a).Return("", errors.New("Boom"))

	resp, err := target.CreateAgent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = failed to create agent 123: Boom", err.Error())
}

func TestCreateAgentMissingRequiredField(t *testing.T) {
	target, _ := SetupTest()
	request := &api.CreateAgentRequest{
		Agent: &api.Agent{
			Id:                  "",
			Name:                "Test Agent",
			AssignedSchemaId:    "",
			EndorsableSchemaIds: nil,
		},
	}

	resp, err := target.CreateAgent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = InvalidArgument desc = name and id are required fields", err.Error())
}

func TestCreateAgentAlreadyExists(t *testing.T) {
	target, suite := SetupTest()
	request := &api.CreateAgentRequest{
		Agent: &api.Agent{
			Id:                  "123",
			Name:                "Test Agent",
			AssignedSchemaId:    "",
			EndorsableSchemaIds: nil,
		},
	}

	suite.Store.On("GetAgent", "123").Return(nil, nil)

	resp, err := target.CreateAgent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = AlreadyExists desc = agent with id 123 already exists", err.Error())
}

func TestGetAgent(t *testing.T) {
	target, suite := SetupTest()
	request := &api.GetAgentRequest{
		Id: "123",
	}

	suite.Store.On("GetAgent", "123").Return(&datastore.Agent{ID: "123", Name: "test Agent"}, nil)

	resp, err := target.GetAgent(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, "test Agent", resp.Agent.Name)
}

func TestGetAgentErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.GetAgentRequest{
		Id: "123",
	}

	suite.Store.On("GetAgent", "123").Return(nil, errors.New("BOOM"))

	resp, err := target.GetAgent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = unable to get agent: BOOM", err.Error())
}

func TestListAgent(t *testing.T) {
	target, suite := SetupTest()
	request := &api.ListAgentRequest{}

	suite.Store.On("ListAgent", &datastore.AgentCriteria{}).Return(&datastore.AgentList{
		Count:  1,
		Agents: []*datastore.Agent{{ID: "123", Name: "test agent"}},
	}, nil)

	resp, err := target.ListAgent(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, "test agent", resp.Agents[0].Name)
}

func TestListAgentErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.ListAgentRequest{}

	suite.Store.On("ListAgent", &datastore.AgentCriteria{}).Return(nil, errors.New("BOOM"))

	resp, err := target.ListAgent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = unable to list agent: BOOM", err.Error())
}

func TestDeleteAgent(t *testing.T) {
	target, suite := SetupTest()
	request := &api.DeleteAgentRequest{
		Id: "123",
	}
	agent := &datastore.Agent{ID: "123"}
	suite.Store.On("GetAgent", "123").Return(agent, nil)
	suite.Store.On("DeleteAgent", "123").Return(nil)

	resp, err := target.DeleteAgent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteAgentErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.DeleteAgentRequest{
		Id: "123",
	}

	agent := &datastore.Agent{ID: "123"}
	suite.Store.On("GetAgent", "123").Return(agent, nil)
	suite.Store.On("DeleteAgent", "123").Return(errors.New("BOOM"))

	resp, err := target.DeleteAgent(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = failed to delete agent 123: BOOM", err.Error())
}

func TestUpdateAgent(t *testing.T) {
	target, suite := SetupTest()
	request := &api.UpdateAgentRequest{
		Agent: &api.Agent{
			Id:                  "123",
			Name:                "Test Agent",
			AssignedSchemaId:    "",
			EndorsableSchemaIds: nil,
		},
	}

	a := &datastore.Agent{
		ID:                  "123",
		Name:                "Test Agent",
		AssignedSchemaId:    "",
		EndorsableSchemaIds: nil,
	}

	suite.Store.On("GetAgent", "123").Return(a, nil)
	suite.Store.On("UpdateAgent", a).Return(nil)

	resp, err := target.UpdateAgent(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestCreateSchema(t *testing.T) {
	target, suite := SetupTest()
	request := &api.CreateSchemaRequest{
		Schema: &api.Schema{
			Id:      "123",
			Name:    "Test Schema",
			Version: "0.0.1",
			Attributes: []*api.Attribute{{
				Name: "City",
				Type: api.Attribute_STRING,
			}},
		},
	}

	s := &datastore.Schema{
		ID:      "123",
		Name:    "Test Schema",
		Version: "0.0.1",
		Attributes: []*datastore.Attribute{{
			Name: "City",
			Type: int32(api.Attribute_STRING),
		}},
	}
	s2 := &datastore.Schema{
		ID:               "123",
		Name:             "Test Schema",
		Version:          "0.0.1",
		ExternalSchemaID: "abc",
		Attributes: []*datastore.Attribute{{
			Name: "City",
			Type: int32(api.Attribute_STRING),
		}},
	}

	suite.Store.On("GetSchema", "123").Return(nil, errors.New("not found"))
	suite.CredRegistry.On("CreateSchema", s).Return("abc", nil)
	suite.Store.On("InsertSchema", s2).Return("123", nil)

	resp, err := target.CreateSchema(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, resp.Id, "123")
}

func TestCreateSchemaFails(t *testing.T) {
	target, suite := SetupTest()
	request := &api.CreateSchemaRequest{
		Schema: &api.Schema{
			Id:   "123",
			Name: "Test Schema",
		},
	}

	s := &datastore.Schema{
		ID:         "123",
		Name:       "Test Schema",
		Attributes: []*datastore.Attribute{},
	}
	s2 := &datastore.Schema{
		ID:               "123",
		Name:             "Test Schema",
		ExternalSchemaID: "abc",
		Attributes:       []*datastore.Attribute{},
	}

	suite.Store.On("GetSchema", "123").Return(nil, errors.New("not found"))
	suite.CredRegistry.On("CreateSchema", s).Return("abc", nil)
	suite.Store.On("InsertSchema", s2).Return("", errors.New("Boom"))

	resp, err := target.CreateSchema(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = failed to create schema 123: Boom", err.Error())
}

func TestCreateSchemaMissingRequiredField(t *testing.T) {
	target, _ := SetupTest()
	request := &api.CreateSchemaRequest{
		Schema: &api.Schema{
			Id:   "",
			Name: "Test Schema",
		},
	}

	resp, err := target.CreateSchema(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = InvalidArgument desc = name and id are required fields", err.Error())
}

func TestCreateSchemaAlreadyExists(t *testing.T) {
	target, suite := SetupTest()
	request := &api.CreateSchemaRequest{
		Schema: &api.Schema{
			Id:   "123",
			Name: "Test Schema",
		},
	}

	suite.Store.On("GetSchema", "123").Return(nil, nil)

	resp, err := target.CreateSchema(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = AlreadyExists desc = schema with id 123 already exists", err.Error())
}

func TestGetSchema(t *testing.T) {
	target, suite := SetupTest()
	request := &api.GetSchemaRequest{
		Id: "123",
	}

	suite.Store.On("GetSchema", "123").Return(&datastore.Schema{
		ID:   "123",
		Name: "test schema",
		Attributes: []*datastore.Attribute{
			{
				Name: "City",
				Type: 1,
			},
		},
	}, nil)

	resp, err := target.GetSchema(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, "test schema", resp.Schema.Name)
}

func TestGetSchemaErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.GetSchemaRequest{
		Id: "123",
	}

	suite.Store.On("GetSchema", "123").Return(nil, errors.New("BOOM"))

	resp, err := target.GetSchema(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = unable to get schema: BOOM", err.Error())
}

func TestListSchema(t *testing.T) {
	target, suite := SetupTest()
	request := &api.ListSchemaRequest{}

	suite.Store.On("ListSchema", &datastore.SchemaCriteria{}).Return(&datastore.SchemaList{
		Count: 1,
		Schema: []*datastore.Schema{{
			ID:   "123",
			Name: "test schema",
			Attributes: []*datastore.Attribute{
				{
					Name: "City",
					Type: 1,
				},
			},
		}},
	}, nil)

	resp, err := target.ListSchema(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, "test schema", resp.Schema[0].Name)
}

func TestListSchemaErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.ListSchemaRequest{}

	suite.Store.On("ListSchema", &datastore.SchemaCriteria{}).Return(nil, errors.New("BOOM"))

	resp, err := target.ListSchema(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = unable to list schema: BOOM", err.Error())
}

func TestDeleteSchema(t *testing.T) {
	target, suite := SetupTest()
	request := &api.DeleteSchemaRequest{
		Id: "123",
	}

	suite.Store.On("DeleteSchema", "123").Return(nil)

	resp, err := target.DeleteSchema(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteSchemaErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.DeleteSchemaRequest{
		Id: "123",
	}

	suite.Store.On("DeleteSchema", "123").Return(errors.New("BOOM"))

	resp, err := target.DeleteSchema(context.Background(), request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "rpc error: code = Internal desc = failed to delete schema 123: BOOM", err.Error())
}

func TestUpdateSchema(t *testing.T) {
	target, suite := SetupTest()
	request := &api.UpdateSchemaRequest{
		Schema: &api.Schema{
			Id:      "123",
			Name:    "Test Schema",
			Version: "0.0.1",
			Attributes: []*api.Attribute{{
				Name: "City",
				Type: api.Attribute_STRING,
			}},
		},
	}

	a := &datastore.Schema{
		ID:      "123",
		Name:    "Test Schema",
		Version: "0.0.1",
		Attributes: []*datastore.Attribute{{
			Name: "City",
			Type: int32(api.Attribute_STRING),
		}},
	}

	suite.Store.On("GetSchema", "123").Return(a, nil)
	suite.Store.On("UpdateSchema", a).Return(nil)

	resp, err := target.UpdateSchema(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestUpdateSchemaErr(t *testing.T) {
	target, suite := SetupTest()
	request := &api.UpdateSchemaRequest{
		Schema: &api.Schema{
			Id:      "123",
			Name:    "Test Schema",
			Version: "0.0.1",
			Attributes: []*api.Attribute{{
				Name: "City",
				Type: api.Attribute_STRING,
			}},
		},
	}

	suite.Store.On("GetSchema", "123").Return(nil, errors.New("BOOM"))

	resp, err := target.UpdateSchema(context.Background(), request)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestUpdateSchemaDataMissing(t *testing.T) {
	target, _ := SetupTest()
	request := &api.UpdateSchemaRequest{
		Schema: &api.Schema{
			Id:      "",
			Name:    "Test Schema",
			Version: "0.0.1",
			Attributes: []*api.Attribute{{
				Name: "City",
				Type: api.Attribute_STRING,
			}},
		},
	}

	resp, err := target.UpdateSchema(context.Background(), request)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}
