/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifier

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	api "github.com/scoir/canis/pkg/didcomm/verifier/api/protogen"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/presentproof"
	"github.com/scoir/canis/pkg/presentproof/engine"
)

type proofClient interface {
	SendRequestPresentation(msg *ppclient.RequestPresentation, myDID, theirDID string) (string, error)
}

type Server struct {
	store        datastore.Store
	proofcl      proofClient
	ctx          *ariescontext.Provider
	ppsup        *presentproof.Supervisor
	registry     engine.PresentationRegistry
	proofHandler *proofHandler
}

type provider interface {
	Store() datastore.Store
	GetAriesContext() (*ariescontext.Provider, error)
	GetPresentationEngineRegistry() (engine.PresentationRegistry, error)
}

func New(ctx provider) (*Server, error) {

	actx, err := ctx.GetAriesContext()
	if err != nil {
		return nil, err
	}

	prov := framework.NewSimpleProvider(actx)
	proofcl, err := prov.GetPresentProofClient()
	if err != nil {
		log.Fatalln("unable to get present proof client")
	}

	reg, err := ctx.GetPresentationEngineRegistry()
	if err != nil {
		log.Fatalln("unable to initialize proof engine registry", err)
	}

	ppsup, err := presentproof.New(prov)
	if err != nil {
		log.Fatalln("unable to create new proof supervisor", err)
	}

	store := ctx.Store()
	handler := &proofHandler{
		ctx:      actx,
		ppsup:    ppsup,
		store:    store,
		registry: reg,
	}
	err = ppsup.Start(handler)
	if err != nil {
		log.Fatalln("unable to start proof supervisor", err)
	}

	r := &Server{
		store:        store,
		proofcl:      proofcl,
		ppsup:        ppsup,
		proofHandler: handler,
		registry:     reg,
	}

	return r, nil
}

func (r *Server) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterVerifierServer(server, r)
}

func (r *Server) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {
	//NO-OP
}

func (r *Server) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *Server) RequestPresentation(_ context.Context, req *api.RequestPresentationRequest) (*api.RequestPresentationResponse, error) {
	agent, err := r.store.GetAgent(req.AgentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to load agent: %v", err))
	}

	ac, err := r.store.GetAgentConnection(agent, req.ExternalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load connection: %v", err))
	}

	schema, err := r.store.GetSchema(req.SchemaId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load schema: %v", err))
	}

	presentation, err := r.registry.RequestPresentation(schema.Type, req.RequestedAttributes, req.RequestedPredicates)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error creating presentation request: %v", err))
	}

	requestPresentationID, err := r.proofcl.SendRequestPresentation(presentation, ac.MyDID, ac.TheirDID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error sending presentation request: %v", err))
	}

	prs := &datastore.PresentationRequest{
		AgentID:               agent.ID,
		SchemaID:              schema.ID,
		ExternalID:            req.ExternalId,
		PresentationRequestID: requestPresentationID,
	}

	id, err := r.store.InsertPresentationRequest(prs)
	if err != nil {
		return nil, err
	}

	return &api.RequestPresentationResponse{RequestPresentationId: id}, nil
}
