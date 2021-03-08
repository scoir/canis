package cloudagent

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ariesdidex "github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	api "github.com/scoir/canis/pkg/didcomm/cloudagent/api/protogen"
	"github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/protogen/common"
)

const (
	CanisCloudAgentIDHeaderKey  = "x-canis-cloud-agent-id"
	CanisCloudAgentSigHeaderKey = "x-canis-cloud-agent-signature"
)

type CloudAgent struct {
	store datastore.Store

	external         string
	vdriReg          vdriapi.Registry
	bouncer          didexchange.Bouncer
	cloudAgentSecret string
	grpcHost         string
	grpcPort         int
	grpcBridgeHost   string
	grpcBridgePort   int
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	GetAriesContext() (*ariescontext.Provider, error)
	GetDatastore() (datastore.Store, error)
	GetCloudAgentSecret() string
	GetExternal() string
	GetGRPCEndpoint() (*framework.Endpoint, error)
	GetBridgeEndpoint() (*framework.Endpoint, error)
}

func New(ctx provider) (*CloudAgent, error) {
	r := &CloudAgent{
		external:         ctx.GetExternal(),
		cloudAgentSecret: ctx.GetCloudAgentSecret(),
	}

	e, err := ctx.GetGRPCEndpoint()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get grpc endpoint")
	}

	r.grpcHost = e.Host
	r.grpcPort = e.Port

	e, err = ctx.GetBridgeEndpoint()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get grpc bridge endpoint")
	}

	r.grpcBridgeHost = e.Host
	r.grpcBridgePort = e.Port

	store, err := ctx.GetDatastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to datastore to start ariesmediator")
	}

	r.store = store

	ap, err := ctx.GetAriesContext()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries contect for ariesmediator")
	}

	simp := framework.NewSimpleProvider(ap)

	bouncer, err := didexchange.NewBouncer(simp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get bouncer")
	}

	r.bouncer = bouncer

	r.vdriReg = ap.VDRIRegistry()

	return r, nil
}

// APISpec is a no-op for an internal component
func (r *CloudAgent) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *CloudAgent) RegisterCloudAgent(_ context.Context, request *common.RegisterCloudAgentRequest) (*common.RegisterCloudAgentResponse, error) {

	if request.Secret != r.cloudAgentSecret {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid edge agent secret"))
	}

	id, err := r.store.RegisterCloudAgent(request.ExternalId, request.PublicKey, request.NextKey)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to register edge agent with ID %s", request.ExternalId))
	}

	out := &common.RegisterCloudAgentResponse{
		CloudAgentId: id,
	}
	return out, nil
}

func (r *CloudAgent) PollConnectionRequests(ctx context.Context, request *common.PollConnectionRequest) (*common.PollConnectionResponse, error) {
	panic("implement me")
}

func (r *CloudAgent) HandleInvitation(ctx context.Context, req *common.HandleInvitationRequest) (*common.HandleInvitationResponse, error) {

	cloudAgentID := r.getAgentID(ctx)
	agent, err := r.store.GetCloudAgent(cloudAgentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("agent with id %s not found", cloudAgentID))
	}

	invite := &ariesdidex.Invitation{}
	err = json.Unmarshal([]byte(req.Invitation), invite)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invitation is not valid JSON for an invite")
	}

	err = r.bouncer.EstablishConnectionNotify(invite, r.accepted(agent), failed)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("error creating invitation for agent %s", agent.ID))
	}

	return &common.HandleInvitationResponse{}, nil
}

func (r *CloudAgent) AcceptConnection(ctx context.Context, request *common.AcceptConnectionRequest) (*common.AcceptConnectionResponse, error) {
	panic("implement me")
}

func (r *CloudAgent) PollCredentialOffers(ctx context.Context, request *common.PollCredentialOffersRequest) (*common.PollCredentialOffersResponse, error) {
	panic("implement me")
}

func (r *CloudAgent) AcceptCredential(ctx context.Context, request *common.AcceptCredentialRequest) (*common.AcceptCredentialResponse, error) {
	panic("implement me")
}

func (r *CloudAgent) ListConnections(ctx context.Context, request *common.ListConnectionsRequest) (*common.ListConnectionsResponse, error) {
	cloudAgentID := r.getAgentID(ctx)

	cloudAgent, err := r.store.GetCloudAgent(cloudAgentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid agent ID")
	}

	connections, err := r.store.ListCloudAgentConnections(cloudAgent)
	if err != nil {
		return nil, status.Error(codes.Internal, "error retrieving connections")
	}

	out := &common.ListConnectionsResponse{
		Count: int64(len(connections)),
	}

	for _, connection := range connections {
		out.Connections = append(out.Connections, &common.Connection{
			Id:       connection.ConnectionID,
			Name:     connection.TheirLabel,
			TheirDid: connection.TheirDID,
			MyDid:    connection.MyDID,
		})
	}

	return out, nil
}

func (r *CloudAgent) ListCredentials(ctx context.Context, request *common.ListCredentialsRequest) (*common.ListCredentialsResponse, error) {
	panic("implement me")
}

func (r *CloudAgent) GetEndpoint(_ context.Context, _ *common.EndpointRequest) (*common.EndpointResponse, error) {
	return &common.EndpointResponse{Endpoint: r.external}, nil
}

func (r *CloudAgent) accepted(cloudAgent *datastore.CloudAgent) func(id string, conn *ariesdidex.Connection) {
	return func(id string, conn *ariesdidex.Connection) {

		err := r.store.InsertCloudAgentConnection(cloudAgent, conn)
		if err != nil {
			log.Println("error updating cloud agent configuration", err)
			return
		}

		log.Printf("Successfully connected to cloud agent %s to connection %s", id, "succeeded!")
	}
}

func (r *CloudAgent) Start() error {
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
			log.Println("webbridge server exited with error", err)
		}
	}()

	wg.Wait()
	return nil
}

func failed(id string, err error) {
	log.Println("Connection to", id, "failed with error:", err)
}

func (r *CloudAgent) launchGRPC() error {
	addr := fmt.Sprintf("%s:%d", r.grpcHost, r.grpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	api.RegisterCloudAgentServer(grpcServer, r)
	log.Println("GRPC Listening on ", addr)
	return grpcServer.Serve(lis)
}

func (r *CloudAgent) launchWebBridge() error {
	rmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(h string) (string, bool) {
			if strings.HasPrefix(h, "X-Canis") {
				return h, true
			}

			return "", false
		}),
		runtime.WithMarshalerOption("image/png", &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{OrigName: true},
		}))
	u := fmt.Sprintf("%s:%d", r.grpcBridgeHost, r.grpcBridgePort)
	if u == ":0" {
		return nil
	}

	endpoint := fmt.Sprintf("%s:%d", r.grpcHost, r.grpcPort)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := api.RegisterCloudAgentHandlerFromEndpoint(context.Background(), rmux, endpoint, opts)
	if err != nil {
		log.Println("unable to register admin gateway", err)
	}

	var mux = http.NewServeMux()
	var h http.Handler = rmux
	h = r.signedTokenAuth(rmux)
	mux.Handle("/", h)

	log.Printf("GRPC Web Gateway listening on %s\n", u)
	return http.ListenAndServe(u, mux)
}

func (r *CloudAgent) signedTokenAuth(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		if req.URL.Path == "/cloudagents" && req.Method == http.MethodPost {
			h.ServeHTTP(w, req)
			return
		}

		cloudAgentID := req.Header.Get(CanisCloudAgentIDHeaderKey)
		if cloudAgentID == "" {
			http.Error(w, "Not authorized", 401)
			return
		}

		sig := req.Header.Get(CanisCloudAgentSigHeaderKey)
		if sig == "" {
			http.Error(w, "Not authorized", 401)
			return
		}

		cloudAgent, err := r.store.GetCloudAgent(cloudAgentID)
		if err != nil {
			log.Println("error loading cloud agent", cloudAgentID, err)
			http.Error(w, "Not authorized", 401)
			return
		}

		var msg []byte
		if req.Method == http.MethodGet || req.Method == http.MethodDelete {
			msg = []byte(req.URL.EscapedPath())
		} else {
			msg, _ = ioutil.ReadAll(req.Body)
			req.Body = ioutil.NopCloser(bytes.NewBuffer(msg))
		}

		bsig, err := base64.URLEncoding.DecodeString(sig)
		if err != nil {
			http.Error(w, "Not authorized", 401)
			return
		}

		if ed25519.Verify(cloudAgent.PublicKey, msg, bsig) {
			h.ServeHTTP(w, req)
			return
		}
		http.Error(w, "Not authorized", 401)
		return
	}
}

func (r *CloudAgent) getAgentID(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	cloudAgentID := md.Get(CanisCloudAgentIDHeaderKey)
	if len(cloudAgentID) != 1 {
		return ""
	}
	return cloudAgentID[0]
}
