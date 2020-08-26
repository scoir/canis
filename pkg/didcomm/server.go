/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didcomm

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/didcomm/api"
)

type Server struct {
	didcl  *didexchange.Client
	credcl *issuecredential.Client
}

type provider interface {
	GetStorageProvider() (storage.Provider, error)
}

func New(ctx provider) (*Server, error) {

	sp, err := ctx.GetStorageProvider()
	if err != nil {
		return nil, err
	}

	f, err := aries.New(
		aries.WithTransportReturnRoute("all"),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithStoreProvider(sp),
	)

	if err != nil {
		return nil, err
	}

	fctx, err := f.Context()
	if err != nil {
		return nil, err
	}

	didcl, err := didexchange.New(fctx)
	if err != nil {
		return nil, err
	}

	credcl, err := issuecredential.New(fctx)
	if err != nil {
		return nil, err
	}

	r := &Server{
		didcl:  didcl,
		credcl: credcl,
	}

	return r, nil
}

func (s Server) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterIssuerServer(server, NewIssuer(s.didcl, s.credcl))
}

func (s Server) GetServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{}
}

func (s Server) RegisterGRPCGateway(mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption) {
	panic("implement me")
}

func (s Server) APISpec() (http.HandlerFunc, error) {
	panic("implement me")
}
