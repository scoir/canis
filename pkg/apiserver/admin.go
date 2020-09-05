/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/datastore"
	doorman "github.com/scoir/canis/pkg/didcomm/doorman/api"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
	"github.com/scoir/canis/pkg/static"
)

func (r *APIServer) RegisterGRPCGateway(mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) {
	err := api.RegisterAdminHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	if err != nil {
		log.Println("unable to register admin gateway", err)
	}
}

func (r *APIServer) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterAdminServer(server, r)
}

func (r *APIServer) APISpec() (http.HandlerFunc, error) {
	return static.ServeHTTP, nil
}

func (r *APIServer) CreateSchema(_ context.Context, req *api.CreateSchemaRequest) (*api.CreateSchemaResponse, error) {
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

func (r *APIServer) ListSchema(_ context.Context, req *api.ListSchemaRequest) (*api.ListSchemaResponse, error) {
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

func (r *APIServer) GetSchema(_ context.Context, req *api.GetSchemaRequest) (*api.GetSchemaResponse, error) {
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

func (r *APIServer) DeleteSchema(_ context.Context, req *api.DeleteSchemaRequest) (*api.DeleteSchemaResponse, error) {
	err := r.schemaStore.DeleteSchema(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to delete schema %s", req.Id).Error())
	}

	return &api.DeleteSchemaResponse{}, nil
}

func (r *APIServer) UpdateSchema(_ context.Context, req *api.UpdateSchemaRequest) (*api.UpdateSchemaResponse, error) {
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

func (r *APIServer) CreateAgent(_ context.Context, req *api.CreateAgentRequest) (*api.CreateAgentResponse, error) {
	a := &datastore.Agent{
		ID:                  req.Agent.Id,
		Name:                req.Agent.Name,
		AssignedSchemaId:    req.Agent.AssignedSchemaId,
		EndorsableSchemaIds: req.Agent.EndorsableSchemaIds,
		HasPublicDID:        req.Agent.PublicDid,
	}
	//TODO:  how do we get this agents Endpoint??

	if a.ID == "" || a.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name and id are required fields")
	}
	_, err := r.agentStore.GetAgent(a.ID)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("agent with id %s already exists", req.Agent.Id))
	}

	if a.HasPublicDID {
		err = r.createAgentWallet(a)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to provision agent wallet %s", req.Agent.Id).Error())
		}
	}

	id, err := r.agentStore.InsertAgent(a)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create agent %s", req.Agent.Id).Error())
	}

	r.fireAgentEvent(&api.AgentEvent{
		Type: api.AgentEvent_ADD,
		New:  req.Agent,
	})

	return &api.CreateAgentResponse{
		Id: id,
	}, nil
}

func (r *APIServer) ListAgent(_ context.Context, req *api.ListAgentRequest) (*api.ListAgentResponse, error) {
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

func (r *APIServer) GetAgent(_ context.Context, req *api.GetAgentRequest) (*api.GetAgentResponse, error) {
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

func (r *APIServer) GetAgentInvitation(ctx context.Context, request *api.InvitationRequest) (*api.InvitationResponse, error) {
	doormanReq := &doorman.InvitationRequest{
		AgentId:    request.AgentId,
		ExternalId: request.ExternalId,
		Name:       request.Name,
	}
	invite, err := r.doorman.GetInvitation(ctx, doormanReq)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to get agent invitation").Error())
	}

	return &api.InvitationResponse{
		Invitation: invite.Invitation,
	}, nil
}

func (r *APIServer) DeleteAgent(_ context.Context, req *api.DeleteAgentRequest) (*api.DeleteAgentResponse, error) {

	old, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, errors.Wrapf(err, "unable to find agent %s to deleteS", req.Id).Error())
	}

	err = r.agentStore.DeleteAgent(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to delete agent %s", req.Id).Error())
	}

	out := &api.Agent{
		Id:                  old.ID,
		Name:                old.Name,
		AssignedSchemaId:    old.AssignedSchemaId,
		EndorsableSchemaIds: old.EndorsableSchemaIds,
	}

	r.fireAgentEvent(&api.AgentEvent{
		Type: api.AgentEvent_DELETE,
		Old:  out,
	})

	return &api.DeleteAgentResponse{}, nil
}

func (r *APIServer) UpdateAgent(_ context.Context, req *api.UpdateAgentRequest) (*api.UpdateAgentResponse, error) {
	if req.Agent.Id == "" || req.Agent.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name and id are required fields")
	}

	old, err := r.agentStore.GetAgent(req.Agent.Id)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("agent with id %s already exists", req.Agent.Id))
	}

	var upd = *old
	upd.Name = req.Agent.Name
	upd.AssignedSchemaId = req.Agent.AssignedSchemaId
	upd.EndorsableSchemaIds = req.Agent.EndorsableSchemaIds

	err = r.agentStore.UpdateAgent(&upd)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create agent %s", req.Agent.Id).Error())
	}

	r.fireAgentEvent(&api.AgentEvent{
		Type: api.AgentEvent_UPDATE,
		Old: &api.Agent{
			Id:                  old.ID,
			Name:                old.Name,
			AssignedSchemaId:    old.AssignedSchemaId,
			EndorsableSchemaIds: old.EndorsableSchemaIds,
		},
		New: &api.Agent{
			Id:                  upd.ID,
			Name:                upd.Name,
			AssignedSchemaId:    upd.AssignedSchemaId,
			EndorsableSchemaIds: upd.EndorsableSchemaIds,
		},
	})

	return &api.UpdateAgentResponse{}, nil
}

func (r *APIServer) WatchAgents(_ *api.WatchRequest, stream api.Admin_WatchAgentsServer) error {

	ch := r.registerWatcher()
	defer r.removeWatcher(ch)
	for ae := range ch {
		if err := stream.Send(ae); err != nil {
			return err
		}
	}

	return nil
}

func (r *APIServer) removeWatcher(ch chan *api.AgentEvent) {
	r.watcherLock.Lock()
	defer r.watcherLock.Unlock()

	close(ch)

	for i, watcher := range r.watchers {
		if watcher == ch {
			copy(r.watchers[i:], r.watchers[i+1:])
			r.watchers[len(r.watchers)-1] = nil
			r.watchers = r.watchers[:len(r.watchers)-1]
			break
		}
	}
}

func (r *APIServer) registerWatcher() chan *api.AgentEvent {
	r.watcherLock.Lock()
	defer r.watcherLock.Unlock()

	ch := make(chan *api.AgentEvent)
	r.watchers = append(r.watchers, ch)

	return ch
}

func (r *APIServer) fireAgentEvent(evt *api.AgentEvent) {
	r.watcherLock.RLock()
	defer r.watcherLock.RUnlock()

	for _, watcher := range r.watchers {
		watcher <- evt
	}
}

func (r *APIServer) SeedPublicDID(_ context.Context, req *api.SeePublicDIDRequest) (*api.SeedPublicDIDResponse, error) {
	_, err := r.didStore.GetPublicDID()
	if err == nil {
		return nil, status.Error(codes.FailedPrecondition, "public DID already exists")
	}

	did, keyPair, err := identifiers.CreateDID(&identifiers.MyDIDInfo{
		Seed:       req.Seed,
		Cid:        true,
		MethodName: "scr",
	})

	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create new DID for apiserver PublicDID").Error())
	}

	_, err = r.client.GetNym(did.String())
	if err != nil {
		log.Fatalln("DID must be registered to be public", err)
	}

	var d = &datastore.DID{
		DID: did,
		KeyPair: &datastore.KeyPair{
			PublicKey:  keyPair.PublicKey(),
			PrivateKey: keyPair.PrivateKey(),
		},
		Endpoint: "",
	}
	err = r.didStore.InsertDID(d)
	if err != nil {
		log.Fatalln(err)
	}

	err = r.didStore.SetPublicDID(did.String())
	if err != nil {
		log.Fatalln(err)
	}

	return &api.SeedPublicDIDResponse{}, nil
}
