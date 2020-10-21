/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/issuer/api"
	"github.com/scoir/canis/pkg/framework"
)

type credentialIssuer interface {
	SendOffer(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error)
}

type Server struct {
	store       datastore.Store
	credcl      credentialIssuer
	ctx         *ariescontext.Provider
	credsup     *credential.Supervisor
	registry    engine.CredentialRegistry
	credHandler *credHandler
}

type provider interface {
	Store() datastore.Store
	GetAriesContext() (*ariescontext.Provider, error)
	GetCredentialEngineRegistry() (engine.CredentialRegistry, error)
}

func New(ctx provider) (*Server, error) {

	actx, err := ctx.GetAriesContext()
	prov := framework.NewSimpleProvider(actx)
	credcl, err := prov.GetCredentialClient()
	if err != nil {
		log.Fatalln("unable to get credential client")
	}

	reg, err := ctx.GetCredentialEngineRegistry()
	if err != nil {
		log.Fatalln("unable to initialize credential engine registry", err)
	}

	credsup, err := credential.New(prov)
	if err != nil {
		log.Fatalln("unable to create new credential supervisor", err)
	}

	store := ctx.Store()
	handler := &credHandler{
		ctx:      actx,
		credsup:  credsup,
		store:    store,
		registry: reg,
	}
	err = credsup.Start(handler)
	if err != nil {
		log.Fatalln("unable to start credential supervisor", err)
	}

	r := &Server{
		store:       store,
		credcl:      credcl,
		credsup:     credsup,
		credHandler: handler,
		registry:    reg,
	}

	return r, nil
}

func (r *Server) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterIssuerServer(server, r)
}

func (r *Server) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {
	//NO-OP
}

func (r *Server) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *Server) IssueCredential(_ context.Context, req *api.IssueCredentialRequest) (*api.IssueCredentialResponse, error) {

	agent, err := r.store.GetAgent(req.AgentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to load agent: %v", err))
	}

	ac, err := r.store.GetAgentConnection(agent, req.SubjectId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load connection: %v", err))
	}

	schema, err := r.store.GetSchema(req.Credential.SchemaId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load schema: %v", err))
	}

	vals := map[string]interface{}{}
	attrs := make([]icprotocol.Attribute, len(req.Credential.Attributes))
	for i, a := range req.Credential.Attributes {
		attrs[i] = icprotocol.Attribute{
			Name:  a.Name,
			Value: a.Value,
		}
		vals[a.Name] = a.Value
	}

	registryOfferID, attachment, err := r.registry.CreateCredentialOffer(agent.PublicDID, ac.TheirDID, schema, vals)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error creating credential offer: %v", err))
	}

	offer := &issuecredential.OfferCredential{
		Comment: req.Credential.Comment,
		CredentialPreview: icprotocol.PreviewCredential{
			Type:       req.Credential.Type,
			Attributes: attrs,
		},
		Formats: []icprotocol.Format{
			{
				AttachID: registryOfferID,
				Format:   schema.Format,
			},
		},
		OffersAttach: []decorator.Attachment{
			{
				ID:   registryOfferID,
				Data: *attachment,
			},
		},
	}

	id, err := r.credcl.SendOffer(offer, ac.MyDID, ac.TheirDID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("send offer failed: %v", err))
	}

	cred := &datastore.Credential{
		AgentID:           agent.ID,
		MyDID:             ac.MyDID,
		TheirDID:          ac.TheirDID,
		OfferID:           id,
		RegistryOfferID:   registryOfferID,
		SchemaID:          schema.ID,
		ExternalSubjectID: req.SubjectId,
		Offer: datastore.Offer{
			Comment:    req.Credential.Comment,
			Type:       req.Credential.Type,
			Attributes: attrs,
		},
		SystemState: "offered",
	}

	credID, err := r.store.InsertCredential(cred)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error saving credential: %v", err))
	}

	return &api.IssueCredentialResponse{
		CredentialId: credID,
	}, nil
}
