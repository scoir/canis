/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"

	api "github.com/scoir/canis/pkg/apiserver/api/protogen"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/protogen/common"
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
		ID:      uuid.New().String(),
		Name:    req.Schema.Name,
		Format:  req.Schema.Format,
		Type:    req.Schema.Type,
		Version: req.Schema.Version,
		Context: req.Schema.Context,
	}

	if s.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is a required field")
	}

	s.Attributes = make([]*datastore.Attribute, len(req.Schema.Attributes))
	for i, attr := range req.Schema.Attributes {
		s.Attributes[i] = &datastore.Attribute{
			Name: attr.Name,
			Type: int32(attr.Type),
		}
	}

	_, err := r.schemaStore.GetSchema(s.Name)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("schema with id %s already exists", req.Schema.Name))
	}

	if id, err := r.schemaRegistry.CreateSchema(s); err == nil {
		s.ExternalSchemaID = id
	} else {
		log.Println("error creating schema", err)
	}

	id, err := r.schemaStore.InsertSchema(s)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create schema %s", req.Schema.Name).Error())
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
			Context:    schema.Context,
			Format:     schema.Format,
			Type:       schema.Type,
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
		Context:    schema.Context,
		Format:     schema.Format,
		Type:       schema.Type,
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
	s.Context = req.Schema.Context
	s.Type = req.Schema.Type
	s.Format = req.Schema.Format
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
		ID:                    uuid.New().String(),
		Name:                  req.Agent.Name,
		EndorsableSchemaNames: []string{},
		HasPublicDID:          req.Agent.PublicDid,
	}

	if a.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is a required field")
	}
	_, err := r.agentStore.GetAgent(a.Name)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("agent with name %s already exists", req.Agent.Name))
	}

	if a.HasPublicDID {
		err = r.createAgentPublicDID(a)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to provision agent wallet %s", req.Agent.Name).Error())
		}

		for _, schemaID := range req.Agent.EndorsableSchemaNames {
			schema, err := r.schemaStore.GetSchema(schemaID)
			if err != nil {
				continue
			}
			err = r.schemaRegistry.RegisterSchema(a.PublicDID, schema)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			a.EndorsableSchemaNames = append(a.EndorsableSchemaNames, schemaID)
		}
	}

	id, err := r.agentStore.InsertAgent(a)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create agent %s", req.Agent.Name).Error())
	}

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
			Id:                    Agent.ID,
			Name:                  Agent.Name,
			EndorsableSchemaNames: Agent.EndorsableSchemaNames,
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
		Id:                    Agent.ID,
		Name:                  Agent.Name,
		EndorsableSchemaNames: Agent.EndorsableSchemaNames,
	}

	return out, nil
}

func (r *APIServer) GetAgentInvitation(ctx context.Context, request *common.InvitationRequest) (*common.InvitationResponse, error) {
	doormanReq := &common.InvitationRequest{
		AgentName:  request.AgentName,
		ExternalId: request.ExternalId,
	}
	invite, err := r.doorman.GetInvitation(ctx, doormanReq)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to get agent invitation").Error())
	}

	return &common.InvitationResponse{
		Invitation: invite.Invitation,
	}, nil
}

func (r *APIServer) DeleteAgent(_ context.Context, req *api.DeleteAgentRequest) (*api.DeleteAgentResponse, error) {

	_, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, errors.Wrapf(err, "unable to find agent %s to deleteS", req.Id).Error())
	}

	err = r.agentStore.DeleteAgent(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to delete agent %s", req.Id).Error())
	}

	return &api.DeleteAgentResponse{}, nil
}

func (r *APIServer) UpdateAgent(_ context.Context, req *api.UpdateAgentRequest) (*api.UpdateAgentResponse, error) {
	if req.Agent.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required fields")
	}

	old, err := r.agentStore.GetAgent(req.Agent.Id)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("agent with id %s already exists", req.Agent.Id))
	}

	var upd = *old

	if req.Agent.Name != "" {
		upd.Name = req.Agent.Name
	}

	upd.HasPublicDID = req.Agent.PublicDid
	if upd.HasPublicDID && upd.PublicDID == nil {
		err := r.createAgentPublicDID(&upd)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error creating agent public DID for %s: %v", req.Agent.Id, err))
		}
	}

	for _, schemaID := range req.Agent.EndorsableSchemaNames {
		schema, err := r.schemaStore.GetSchema(schemaID)
		if err != nil {
			continue
		}
		err = r.schemaRegistry.RegisterSchema(upd.PublicDID, schema)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		upd.EndorsableSchemaNames = append(upd.EndorsableSchemaNames, schemaID)
	}

	err = r.agentStore.UpdateAgent(&upd)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create agent %s", req.Agent.Id).Error())
	}

	return &api.UpdateAgentResponse{}, nil
}

func (r *APIServer) SeedPublicDID(_ context.Context, req *api.SeedPublicDIDRequest) (*api.SeedPublicDIDResponse, error) {
	_, err := r.store.GetPublicDID()
	if err == nil {
		return nil, status.Error(codes.FailedPrecondition, "public DID already exists")
	}

	edseed, err := identifiers.ConvertSeed(req.Seed)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to convert seed for apiserver PublicDID").Error())
	}

	var pubkey ed25519.PublicKey
	var privkey ed25519.PrivateKey
	if len(req.Seed) == 0 {
		pubkey, privkey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create new DID for apiserver PublicDID").Error())
		}
	} else if len(req.Seed) == ed25519.SeedSize {
		privkey = ed25519.NewKeyFromSeed(edseed)
		pubkey = privkey.Public().(ed25519.PublicKey)
	} else {
		return nil, status.Error(codes.InvalidArgument, "invalid seed")
	}

	did, err := identifiers.CreateDID(&identifiers.MyDIDInfo{
		PublicKey:  pubkey,
		Cid:        true,
		MethodName: "sov",
	})
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to create new DID for apiserver PublicDID").Error())
	}

	_, err = r.client.GetNym(did.String())
	if err != nil {
		return nil, errors.Wrap(err, "DID must be registered")
	}

	encPubKey := base58.Encode(pubkey)
	recKID, err := localkms.CreateKID(pubkey, kms.ED25519Type)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create KID for public key")
	}
	kid, _, err := r.keyMgr.ImportPrivateKey(privkey, kms.ED25519Type, kms.WithKeyID(recKID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to import private key")
	}

	var d = &datastore.DID{
		DID: did,
		KeyPair: &datastore.KeyPair{
			ID:        kid,
			PublicKey: encPubKey,
		},
		Endpoint: "",
	}

	err = r.store.SetPublicDID(d)
	if err != nil {
		return nil, errors.Wrap(err, "unable to save public DID")
	}

	return &api.SeedPublicDIDResponse{}, nil
}

func (r *APIServer) IssueCredential(ctx context.Context, req *common.IssueCredentialRequest) (*common.IssueCredentialResponse, error) {

	issuerCred := common.Credential{
		SchemaId: req.Credential.SchemaId,
		Comment:  req.Credential.Comment,
		Type:     req.Credential.Type,
		Preview:  make([]*common.CredentialAttribute, len(req.Credential.Preview)),
		Body:     req.Credential.Body,
	}

	for i, attr := range req.Credential.Preview {
		issuerCred.Preview[i] = &common.CredentialAttribute{
			Name:  attr.Name,
			Value: attr.Value,
		}
	}

	issuerReq := &common.IssueCredentialRequest{
		AgentName:  req.AgentName,
		ExternalId: req.ExternalId,
		Credential: &issuerCred,
	}
	issuerResp, err := r.issuer.IssueCredential(ctx, issuerReq)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to initiate credential offer").Error())
	}

	return &common.IssueCredentialResponse{
		CredentialId: issuerResp.CredentialId,
	}, nil
}

func (r *APIServer) RequestPresentation(ctx context.Context, req *common.RequestPresentationRequest) (*common.RequestPresentationResponse, error) {

	pp := make(map[string]*common.AttrInfo)
	for k, v := range req.Presentation.RequestedAttributes {
		pp[k] = &common.AttrInfo{
			Name:         v.Name,
			Restrictions: v.Restrictions,
			NonRevoked:   v.NonRevoked,
		}
	}

	pq := make(map[string]*common.PredicateInfo)
	for k, v := range req.Presentation.RequestedPredicates {
		pq[k] = &common.PredicateInfo{
			Name:         v.Name,
			PType:        v.PType,
			PValue:       v.PValue,
			Restrictions: v.Restrictions,
			NonRevoked:   v.NonRevoked,
		}
	}

	rpr := &common.RequestPresentationRequest{
		AgentName:  req.AgentName,
		ExternalId: req.ExternalId,
		Presentation: &common.RequestPresentation{
			SchemaId:            req.Presentation.SchemaId,
			Comment:             req.Presentation.Comment,
			WillConfirm:         req.Presentation.WillConfirm,
			RequestedAttributes: pp,
			RequestedPredicates: pq,
		},
	}

	resp, err := r.verifier.RequestPresentation(ctx, rpr)
	if err != nil {
		return nil, err
	}

	return &common.RequestPresentationResponse{
		RequestPresentationId: resp.RequestPresentationId,
	}, nil
}

func (r *APIServer) CreateWebhook(_ context.Context, request *api.CreateWebhookRequest) (*api.CreateWebhookResponse, error) {

	for _, webhook := range request.Webhook {

		hook := &datastore.Webhook{
			Type: request.Id,
			URL:  webhook.Url,
		}
		err := r.store.AddWebhook(hook)
		if err != nil {
			return nil, status.Error(codes.Internal, "unexpected error adding webhook")
		}
	}
	return &api.CreateWebhookResponse{}, nil
}

func (r *APIServer) ListWebhook(_ context.Context, request *api.ListWebhookRequest) (*api.ListWebhookResponse, error) {

	hooks, err := r.store.ListWebhooks(request.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error listing webhooks")
	}

	out := make([]*api.Webhook, len(hooks))
	for i, h := range hooks {
		out[i] = &api.Webhook{
			Url: h.URL,
		}
	}

	return &api.ListWebhookResponse{
		Hooks: out,
	}, nil
}

func (r *APIServer) DeleteWebhook(_ context.Context, request *api.DeleteWebhookRequest) (*api.DeleteWebhookResponse, error) {
	err := r.store.DeleteWebhook(request.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to delete webhook for %s: (%v)", request.Id, err))
	}

	return &api.DeleteWebhookResponse{}, nil
}

func (r *APIServer) ListConnections(_ context.Context, req *api.ListConnectionRequest) (*api.ListConnectionResponse, error) {
	agent, err := r.agentStore.GetAgent(req.AgentName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("agent with id %s not found", req.AgentName))
	}

	connections, err := r.store.ListAgentConnections(agent)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("no connections found for agent with id %s", req.AgentName))
	}

	out := make([]*api.Connection, len(connections))
	for i, connection := range connections {
		out[i] = &api.Connection{
			TheirLabel:   connection.TheirLabel,
			MyLabel:      connection.MyLabel,
			TheirDid:     connection.TheirDID,
			MyDid:        connection.MyDID,
			ConnectionId: connection.ConnectionID,
			ExternalId:   connection.ExternalID,
		}
	}

	return &api.ListConnectionResponse{
		Connections: out,
	}, nil

}

func (r *APIServer) DeleteConnection(_ context.Context, req *api.DeleteConnectionRequest) (*api.DeleteConnectionResponse, error) {
	agent, err := r.agentStore.GetAgent(req.AgentName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("agent with id %s not found", req.AgentName))
	}

	err = r.store.DeleteAgentConnection(agent, req.ExternalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to delete connection %s for agent with name %s", req.ExternalId, req.AgentName))
	}

	return &api.DeleteConnectionResponse{}, nil
}

func (r *APIServer) AcceptInvitation(ctx context.Context, req *common.AcceptInvitationRequest) (*common.AcceptInvitationResponse, error) {
	resp, err := r.doorman.AcceptInvitation(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "unable to accept agent invitation").Error())
	}

	return resp, nil
}
