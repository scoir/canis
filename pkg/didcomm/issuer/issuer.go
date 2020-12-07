/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
	api "github.com/scoir/canis/pkg/didcomm/issuer/api/protogen"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/protogen/common"
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
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries context")
	}

	prov := framework.NewSimpleProvider(actx)
	credcl, err := prov.GetCredentialClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get credential client")
	}

	reg, err := ctx.GetCredentialEngineRegistry()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get credential engine registry")
	}

	credsup, err := credential.New(prov)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get credential supervisor")
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
		return nil, errors.Wrap(err, "unable to start supervisor")
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

func (r *Server) IssueCredential(_ context.Context, req *common.IssueCredentialRequest) (*common.IssueCredentialResponse, error) {

	agent, err := r.store.GetAgent(req.AgentName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to load agent: %v", err))
	}

	ac, err := r.store.GetAgentConnection(agent, req.ExternalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load connection: %v", err))
	}

	schema, err := r.store.GetSchema(req.Credential.SchemaId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load schema: %v", err))
	}

	vals := map[string]interface{}{}
	attrs := make([]icprotocol.Attribute, len(req.Credential.Preview))
	for i, a := range req.Credential.Preview {
		attrs[i] = icprotocol.Attribute{
			Name:  a.Name,
			Value: a.Value,
		}
		vals[a.Name] = a.Value
	}

	body, err := req.Credential.Body.MarshalJSON()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error unmarshaling credential body: %v", err))
	}

	registryOfferID, attachment, err := r.registry.CreateCredentialOffer(agent.PublicDID, ac.TheirDID, schema, body)
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

	cred := &datastore.IssuedCredential{
		AgentName:         agent.Name,
		MyDID:             ac.MyDID,
		TheirDID:          ac.TheirDID,
		ThreadID:          id,
		RegistryOfferID:   registryOfferID,
		SchemaName:        schema.Name,
		ExternalSubjectID: req.ExternalId,
		Offer: &datastore.Offer{
			Comment: req.Credential.Comment,
			Type:    req.Credential.Type,
			Preview: attrs,
			Body:    body,
		},
		SystemState: "offered",
	}

	credID, err := r.store.InsertCredential(cred)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error saving credential: %v", err))
	}

	return &common.IssueCredentialResponse{
		CredentialId: credID,
	}, nil
}
