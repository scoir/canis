package doorman

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ariesdidex "github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/doorman/api"
	"github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
)

type provider interface {
	GetAriesContext() (*ariescontext.Provider, error)
	GetDatastore() (datastore.Store, error)
}

type Doorman struct {
	agentStore datastore.Store
	didcl      *ariesdidex.Client
	bouncer    didexchange.Bouncer
}

func New(prov provider) (*Doorman, error) {

	ctx, err := prov.GetAriesContext()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load aries context")
	}

	simp := framework.NewSimpleProvider(ctx)
	bouncer, err := didexchange.NewBouncer(simp)

	agentStore, err := prov.GetDatastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get datastore provider")
	}

	return &Doorman{
		agentStore: agentStore,
		bouncer:    bouncer,
	}, nil
}

func (r *Doorman) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterDoormanServer(server, r)
}

func (r *Doorman) GetServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{}
}

func (r *Doorman) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {

}

func (r *Doorman) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *Doorman) GetInvitation(_ context.Context, request *api.InvitationRequest) (*api.InvitationResponse, error) {

	agent, err := r.agentStore.GetAgent(request.AgentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("agent with id %s not found", request.AgentId))
	}

	invite, err := r.bouncer.CreateInvitationNotify(request.Name, r.accepted(agent, request.ExternalId), failed)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("schema with id %s already exists", request.AgentId))
	}

	d, _ := json.MarshalIndent(invite, " ", " ")
	return &api.InvitationResponse{
		Invitation: string(d),
	}, nil
}

func (r *Doorman) accepted(agent *datastore.Agent, externalID string) func(id string, conn *ariesdidex.Connection) {
	return func(id string, conn *ariesdidex.Connection) {
		err := r.agentStore.InsertAgentConnection(agent, externalID, conn)
		if err != nil {
			log.Println("error creating agent connection")
		}

		log.Printf("Successfully connected agent %s to connection %s", id, "succeeded!")
	}
}

func failed(id string, err error) {
	log.Println("Connection to", id, "failed with error:", err)
}
