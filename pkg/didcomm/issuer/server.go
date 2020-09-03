/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"errors"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/didcomm/issuer/api"
	"github.com/scoir/canis/pkg/framework"
)

type Server struct {
	credcl      *issuecredential.Client
	ctx         *ariescontext.Provider
	credsup     *credential.Supervisor
	credHandler *credHandler
}

type provider interface {
	GetAriesContext() (*ariescontext.Provider, error)
}

func New(ctx provider) (*Server, error) {

	actx, err := ctx.GetAriesContext()
	prov := framework.NewSimpleProvider(actx)
	credcl, _ := prov.GetCredentialClient()

	credsup, err := credential.New(prov)
	if err != nil {
		log.Fatalln("unable to create new credential supervisor", err)
	}

	handler := &credHandler{
		ctx:     actx,
		credsup: credsup,
	}
	err = credsup.Start(handler)
	if err != nil {
		log.Fatalln("unable to start credential supervisor", err)
	}

	r := &Server{
		credcl:      credcl,
		credsup:     credsup,
		credHandler: handler,
	}

	return r, nil
}

func (r *Server) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterIssuerServer(server, NewIssuer(r.credHandler, r.credcl))
}

func (r *Server) GetServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{}
}

func (r *Server) RegisterGRPCGateway(_ *runtime.ServeMux, _ /*endpoint*/ string, _ ...grpc.DialOption) {
	//NO-OP
}

func (r *Server) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}
