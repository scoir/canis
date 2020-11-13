/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package controller

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/goji/httpauth"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/framework"
)

const (
	SecurityRealm    = "Restricted"
	APIKeyHeaderName = "X-API-Key"
)

type AgentController interface {
	RegisterGRPCHandler(server *grpc.Server)
	RegisterGRPCGateway(mux *runtime.ServeMux, endpoint string, opts ...grpc.DialOption)
	APISpec() (http.HandlerFunc, error)
}

type Runner struct {
	ac                       AgentController
	grpcBridgeHost, grpcHost string
	grpcBridgePort, grpcPort int
	swaggerUsername          string
	swaggerPassword          string
	apiToken                 string
	debug                    bool
}

type provider interface {
	GetGRPCEndpoint() (*framework.Endpoint, error)
	GetBridgeEndpoint() (*framework.Endpoint, error)
}

func New(ctx provider, ac AgentController) (*Runner, error) {
	grpce, err := ctx.GetGRPCEndpoint()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create controller")
	}

	var grpcBridgeHost, swaggerUserName, swaggerPassword, apiToken string
	var grpcBridgePort int
	bridge, err := ctx.GetBridgeEndpoint()
	if err == nil {
		grpcBridgeHost = bridge.Host
		grpcBridgePort = bridge.Port
		swaggerUserName = bridge.Username
		swaggerPassword = bridge.Password
		apiToken = bridge.Token
	}

	r := &Runner{
		ac:              ac,
		grpcHost:        grpce.Host,
		grpcPort:        grpce.Port,
		grpcBridgeHost:  grpcBridgeHost,
		grpcBridgePort:  grpcBridgePort,
		swaggerUsername: swaggerUserName,
		swaggerPassword: swaggerPassword,
		apiToken:        apiToken,

		debug: false,
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

	grpcServer := grpc.NewServer()
	r.ac.RegisterGRPCHandler(grpcServer)
	log.Println("GRPC Listening on ", addr)
	return grpcServer.Serve(lis)
}

func (r *Runner) launchWebBridge() error {
	rmux := runtime.NewServeMux()
	u := fmt.Sprintf("%s:%d", r.grpcBridgeHost, r.grpcBridgePort)
	if u == ":0" {
		return nil
	}

	endpoint := fmt.Sprintf("%s:%d", r.grpcHost, r.grpcPort)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	r.ac.RegisterGRPCGateway(rmux, endpoint, opts...)

	fs := http.FileServer(http.Dir("./static/swaggerui"))

	var mux = http.NewServeMux()
	specFunc, err := r.ac.APISpec()
	if err == nil {
		mux.Handle("/spec/", http.StripPrefix("/spec/", specFunc))
	}
	basicOpts := httpauth.AuthOptions{
		Realm:    SecurityRealm,
		User:     r.swaggerUsername,
		Password: r.swaggerPassword,
	}

	basicAuth := httpauth.BasicAuth(basicOpts)

	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))

	var h http.Handler = rmux
	if r.apiToken != "" {
		h = r.basicTokenAuth(rmux)
	}
	mux.Handle("/", h)

	log.Printf("GRPC Web Gateway listening on %s\n", u)
	return http.ListenAndServe(u, mux)
}

func (r *Runner) basicTokenAuth(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get(APIKeyHeaderName)
		if authHeader == "" {
			http.Error(w, "Not authorized", 401)
			return
		}

		givenToken := sha256.Sum256([]byte(authHeader))
		requiredToken := sha256.Sum256([]byte(r.apiToken))

		if subtle.ConstantTimeCompare(givenToken[:], requiredToken[:]) != 1 {
			http.Error(w, "Not authorized", 401)
			return
		}

		h.ServeHTTP(w, req)
	}
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
