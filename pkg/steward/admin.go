/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/static"
	"github.com/scoir/canis/pkg/steward/api"
)

func (r *Steward) RegisterGRPCGateway(mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) {
	err := api.RegisterAdminHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	if err != nil {
		log.Println("unable to register admin gateway", err)
	}
}

func (r *Steward) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterAdminServer(server, r)
}

func (r *Steward) APISpec() (http.HandlerFunc, error) {
	return static.ServeHTTP, nil
}

func (r *Steward) CreateSchema(_ context.Context, req *api.CreateSchemaRequest) (*api.CreateSchemaResponse, error) {
	s := &datastore.Schema{
		ID:      req.Schema.Id,
		Name:    req.Schema.Name,
		Version: req.Schema.Version,
	}

	if s.ID == "" || s.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name and id are required fields")
	}

	s.Attributes = make([]*datastore.Attribute, len(req.Schema.Attributes))
	for i, attr := range req.Schema.Attributes {
		s.Attributes[i] = &datastore.Attribute{
			Name: attr.Name,
			Type: int32(attr.Type),
		}
	}

	_, err := r.schemaStore.GetSchema(s.ID)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("schema with id %s already exists", req.Schema.Id))
	}

	id, err := r.schemaStore.InsertSchema(s)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create schema %s", req.Schema.Id).Error())
	}

	return &api.CreateSchemaResponse{
		Id: id,
	}, nil
}

func (r *Steward) ListSchema(_ context.Context, req *api.ListSchemaRequest) (*api.ListSchemaResponse, error) {
	critter := &datastore.SchemaCriteria{
		Start:    int(req.Start),
		PageSize: int(req.PageSize),
		Name:     req.Name,
	}

	results, err := r.schemaStore.ListSchema(critter)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to list schema").Error())
	}

	out := &api.ListSchemaResponse{
		Count:  int64(results.Count),
		Schema: make([]*api.Schema, len(results.Schema)),
	}

	for i, schema := range results.Schema {
		out.Schema[i] = &api.Schema{
			Id:         schema.ID,
			Name:       schema.Name,
			Version:    schema.Version,
			Attributes: make([]*api.Attribute, len(schema.Attributes)),
		}

		for x, attribute := range schema.Attributes {
			out.Schema[i].Attributes[x] = &api.Attribute{
				Name: attribute.Name,
				Type: api.Attribute_Type(attribute.Type),
			}
		}
	}

	return out, nil
}

func (r *Steward) GetSchema(_ context.Context, req *api.GetSchemaRequest) (*api.GetSchemaResponse, error) {
	schema, err := r.schemaStore.GetSchema(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to get schema").Error())
	}

	out := &api.GetSchemaResponse{}

	out.Schema = &api.Schema{
		Id:         schema.ID,
		Name:       schema.Name,
		Version:    schema.Version,
		Attributes: make([]*api.Attribute, len(schema.Attributes)),
	}

	for x, attribute := range schema.Attributes {
		out.Schema.Attributes[x] = &api.Attribute{
			Name: attribute.Name,
			Type: api.Attribute_Type(attribute.Type),
		}
	}

	return out, nil
}

func (r *Steward) DeleteSchema(_ context.Context, req *api.DeleteSchemaRequest) (*api.DeleteSchemaResponse, error) {
	err := r.schemaStore.DeleteSchema(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to delete schema %s", req.Id).Error())
	}

	return &api.DeleteSchemaResponse{}, nil
}

func (r *Steward) UpdateSchema(_ context.Context, req *api.UpdateSchemaRequest) (*api.UpdateSchemaResponse, error) {
	if req.Schema.Id == "" || req.Schema.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name and id are required fields")
	}

	s, err := r.schemaStore.GetSchema(req.Schema.Id)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("schema with id %s already exists", req.Schema.Id))
	}

	s.Name = req.Schema.Name
	s.Version = req.Schema.Version
	s.Attributes = make([]*datastore.Attribute, len(req.Schema.Attributes))
	for i, attr := range req.Schema.Attributes {
		s.Attributes[i] = &datastore.Attribute{
			Name: attr.Name,
			Type: int32(attr.Type),
		}
	}

	err = r.schemaStore.UpdateSchema(s)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create schema %s", req.Schema.Id).Error())
	}

	return &api.UpdateSchemaResponse{}, nil
}

func (r *Steward) CreateAgent(_ context.Context, req *api.CreateAgentRequest) (*api.CreateAgentResponse, error) {
	a := &datastore.Agent{
		ID:                  req.Agent.Id,
		Name:                req.Agent.Name,
		AssignedSchemaId:    req.Agent.AssignedSchemaId,
		EndorsableSchemaIds: req.Agent.EndorsableSchemaIds,
	}

	if a.ID == "" || a.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name and id are required fields")
	}
	_, err := r.agentStore.GetAgent(a.ID)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("agent with id %s already exists", req.Agent.Id))
	}

	id, err := r.agentStore.InsertAgent(a)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create agent %s", req.Agent.Id).Error())
	}

	return &api.CreateAgentResponse{
		Id: id,
	}, nil
}

func (r *Steward) ListAgent(_ context.Context, req *api.ListAgentRequest) (*api.ListAgentResponse, error) {
	critter := &datastore.AgentCriteria{
		Start:    int(req.Start),
		PageSize: int(req.PageSize),
		Name:     req.Name,
	}

	results, err := r.agentStore.ListAgent(critter)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to list agent").Error())
	}

	out := &api.ListAgentResponse{
		Count:  int64(results.Count),
		Agents: make([]*api.Agent, len(results.Agents)),
	}

	for i, Agent := range results.Agents {
		out.Agents[i] = &api.Agent{
			Id:                  Agent.ID,
			Name:                Agent.Name,
			AssignedSchemaId:    Agent.AssignedSchemaId,
			EndorsableSchemaIds: Agent.EndorsableSchemaIds,
		}
	}

	return out, nil
}

func (r *Steward) GetAgent(_ context.Context, req *api.GetAgentRequest) (*api.GetAgentResponse, error) {
	Agent, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to get agent").Error())
	}

	out := &api.GetAgentResponse{}

	out.Agent = &api.Agent{
		Id:                  Agent.ID,
		Name:                Agent.Name,
		AssignedSchemaId:    Agent.AssignedSchemaId,
		EndorsableSchemaIds: Agent.EndorsableSchemaIds,
	}

	return out, nil
}

func (r *Steward) DeleteAgent(_ context.Context, req *api.DeleteAgentRequest) (*api.DeleteAgentResponse, error) {
	err := r.agentStore.DeleteAgent(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to delete agent %s", req.Id).Error())
	}

	return &api.DeleteAgentResponse{}, nil
}

func (r *Steward) UpdateAgent(_ context.Context, req *api.UpdateAgentRequest) (*api.UpdateAgentResponse, error) {
	if req.Agent.Id == "" || req.Agent.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name and id are required fields")
	}

	s, err := r.agentStore.GetAgent(req.Agent.Id)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("agent with id %s already exists", req.Agent.Id))
	}

	s.Name = req.Agent.Name
	s.AssignedSchemaId = req.Agent.AssignedSchemaId
	s.EndorsableSchemaIds = req.Agent.EndorsableSchemaIds

	err = r.agentStore.UpdateAgent(s)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create agent %s", req.Agent.Id).Error())
	}

	return &api.UpdateAgentResponse{}, nil
}

func (r *Steward) LaunchAgent(_ context.Context, req *api.LaunchAgentRequest) (*api.LaunchAgentResponse, error) {
	agent, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to load agent to launch: %v", err)
	}

	pID, err := r.exec.LaunchAgent(agent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to launch agent: %v", err)
	}
	agent.PID = pID

	err = r.agentStore.UpdateAgent(agent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to save agent: %v", err)
	}

	out := &api.LaunchAgentResponse{
		Status: api.Agent_STARTING,
	}
	if req.Wait {
		w, err := r.exec.Watch(agent.PID)
		if err != nil {
			log.Println("error watching agent")
		}
		stopper := time.AfterFunc(time.Minute, func() {
			w.Stop()
		})
		defer stopper.Stop()

		for event := range w.ResultChan() {
			switch event.RuntimeContext.Status() {
			case datastore.Running:
				out.Status = api.Agent_RUNNING
				return out, nil
			case datastore.Error:
				out.Status = api.Agent_ERROR
				return out, nil
			case datastore.Completed:
				out.Status = api.Agent_TERMINATED
				return out, nil
			}
		}
	}

	return out, nil
}

func (r *Steward) ShutdownAgent(_ context.Context, req *api.ShutdownAgentRequest) (*api.ShutdownAgentResponse, error) {
	agent, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to load agent to shutdown: %v", err)
	}

	if agent.PID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "agent with ID %s is not currently running", req.Id)
	}

	err = r.exec.ShutdownAgent(agent.PID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to shutdown agent: %v", err)
	}

	agent.PID = ""
	err = r.agentStore.UpdateAgent(agent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to save agent after shutdown: %v", err)
	}

	return &api.ShutdownAgentResponse{}, nil
}

func (r *Steward) GetInvitationForAgent(_ context.Context, req *api.AgentInvitiationRequest) (*api.AgentInivitationResponse, error) {

	agent, err := r.agentStore.GetAgent(req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error querying high school for invite: (%v)", err)
	}

	invite, err := r.bouncer.CreateInvitationWithDIDNotify(agent.Name, r.publicDID.DID, r.handleAgentConnection, r.failedConnectionHandler)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create invitation to high school agent: %v", err)
	}

	d, _ := json.Marshal(invite)
	return &api.AgentInivitationResponse{Body: string(d)}, nil
}
