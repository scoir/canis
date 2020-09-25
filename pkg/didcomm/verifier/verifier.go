/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifier

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/verifier/api"
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

func (r *Server) GetServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{}
}

func (r *Server) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {
	//NO-OP
}

func (r *Server) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *Server) RequestPresentation(_ context.Context, req *api.RequestPresentationRequest) (*api.RequestPresentationResponse, error) {
	_, _ = r.proofcl.SendRequestPresentation(nil, "", "")

	return &api.RequestPresentationResponse{
	}, nil
}
