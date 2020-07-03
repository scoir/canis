/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package controller

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

type AgentController interface {
	RegisterGRPCHandler(server *grpc.Server)
	GetServerOpts() []grpc.ServerOption
	RegisterGRPCGateway(mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption)
	APISpec() (http.HandlerFunc, error)
}

type Runner struct {
	ctx                      *context.Provider
	ac                       AgentController
	grpcBridgeHost, grpcHost string
	grpcBridgePort, grpcPort int
	debug                    bool
}

func New(ctx *context.Provider, grpcHost string, grpcPort int, grpcBridgeHost string, grpcBridgePort int, ac AgentController) (*Runner, error) {
	r := &Runner{
		ctx:            ctx,
		ac:             ac,
		grpcHost:       grpcHost,
		grpcPort:       grpcPort,
		grpcBridgeHost: grpcBridgeHost,
		grpcBridgePort: grpcBridgePort,
		debug:          false,
	}

	return r, nil
}

func (r *Runner) Launch() error {

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := r.launchGRPC()
		if err != nil {
			log.Println("grpc server exited with error: ", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := r.launchWebBridge()
		if err != nil {
			log.Println("webhooks server exited with error", err)
		}
	}()

	wg.Wait()
	return nil
}

func (r *Runner) launchGRPC() error {
	addr := fmt.Sprintf("%s:%d", r.grpcHost, r.grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(r.ac.GetServerOpts()...)
	r.ac.RegisterGRPCHandler(grpcServer)
	log.Println("GRPC Listening for on ", addr)
	return grpcServer.Serve(lis)
}

func (r *Runner) launchWebBridge() error {
	rmux := runtime.NewServeMux()
	endpoint := fmt.Sprintf("%s:%d", r.grpcHost, r.grpcPort)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	r.ac.RegisterGRPCGateway(rmux, endpoint, opts...)

	fs := http.FileServer(http.Dir("./static/swaggerui"))

	var mux = http.NewServeMux()
	specFunc, err := r.ac.APISpec()
	if err == nil {
		mux.Handle("/spec/", http.StripPrefix("/spec/", specFunc))
	}
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))
	mux.Handle("/", rmux)

	u := fmt.Sprintf("%s:%d", r.grpcBridgeHost, r.grpcBridgePort)
	log.Printf("grpc web gateway listening on %s\n", u)
	return http.ListenAndServe(u, mux)
}

func CorsHandler() func(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "PATCH", "POST", "DELETE"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Authentication", "Authorization", "Accept",
			"If-Modified-Since", "Cache-Control", "Pragma", "Upgrade", "Connection"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type", "Cache-Control", "Last-Modified", "Upgrade", "Connection"},
		AllowCredentials: true,
	})
	return c.Handler
}

func Logger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		for key, val := range r.Header {
			log.Println(key, ":", val)
		}

		h.ServeHTTP(w, r)

	}

	return http.HandlerFunc(fn)
}
