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

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	api "github.com/scoir/canis/pkg/didcomm/verifier/api/protogen"
	"github.com/scoir/canis/pkg/presentproof/engine"
	"github.com/scoir/canis/pkg/protogen/common"
)

type Server struct {
	store    datastore.Store
	proofcl  PresentProofClient
	ctx      *ariescontext.Provider
	registry engine.PresentationRegistry
}

func New(ctx Provider) (*Server, error) {

	proofcl, err := ctx.GetPresentProofClient()
	if err != nil {
		log.Fatalln("unable to get present proof client")
	}

	reg, err := ctx.GetPresentationEngineRegistry()
	if err != nil {
		log.Fatalln("unable to initialize proof engine registry", err)
	}
	store := ctx.Store()

	r := &Server{
		store:    store,
		proofcl:  proofcl,
		registry: reg,
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

func (r *Server) RequestPresentation(_ context.Context, req *common.RequestPresentationRequest) (*common.RequestPresentationResponse, error) {
	agent, err := r.store.GetAgent(req.AgentName)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to load agent: %v", err))
	}

	ac, err := r.store.GetAgentConnection(agent, req.ExternalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load connection: %v", err))
	}

	var definitions = &presexch.PresentationDefinitions{
		Name:             req.Presentation.Name,
		Purpose:          req.Presentation.Purpose,
		InputDescriptors: make([]*presexch.InputDescriptor, len(req.Presentation.InputDescriptors)),
	}

	for i, descriptor := range req.Presentation.InputDescriptors {
		definitions.InputDescriptors[i] = &presexch.InputDescriptor{
			ID: descriptor.Id,
			Schema: &presexch.Schema{
				URI:     descriptor.Schema.Uri,
				Name:    descriptor.Schema.Name,
				Purpose: descriptor.Schema.Purpose,
			},
		}
	}

	presentation, err := r.registry.RequestPresentation(req.Presentation.Name, req.Presentation.Format, definitions)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error creating presentation request: %v", err))
	}

	attachID := uuid.New().String()
	sendReq := &ppclient.RequestPresentation{
		Formats: []ppprotocol.Format{{
			AttachID: attachID,
			Format:   req.Presentation.Format,
		}},
		RequestPresentationsAttach: []decorator.Attachment{
			{
				ID:       attachID,
				MimeType: "application/json",
				Data:     *presentation,
			},
		},
	}

	requestPresentationID, err := r.proofcl.SendRequestPresentation(sendReq, ac.MyDID, ac.TheirDID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error sending presentation request: %v", err))
	}

	data, _ := presentation.Fetch()
	prs := &datastore.PresentationRequest{
		AgentID:               agent.Name,
		ExternalID:            req.ExternalId,
		PresentationRequestID: requestPresentationID,
		Data:                  data,
	}

	id, err := r.store.InsertPresentationRequest(prs)
	if err != nil {
		return nil, err
	}

	return &common.RequestPresentationResponse{RequestPresentationId: id}, nil
}
