package mediator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ariesdidex "github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ariesmediator "github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	api "github.com/scoir/canis/pkg/didcomm/mediator/api/protogen"
	"github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/protogen/common"
)

type Mediator struct {
	store datastore.Store

	external           string
	vdriReg            vdriapi.Registry
	bouncer            didexchange.Bouncer
	mediatorActionChan <-chan service.DIDCommAction
	client             *ariesmediator.Client
	edgeAgentSecret    string
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	GetAriesContext() (*ariescontext.Provider, error)
	GetDatastore() (datastore.Store, error)
	GetEdgeAgentSecret() string
	GetExternal() string
}

func New(ctx provider) (*Mediator, error) {
	m := &Mediator{
		external:        ctx.GetExternal(),
		edgeAgentSecret: ctx.GetEdgeAgentSecret(),
	}

	store, err := ctx.GetDatastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to datastore to start ariesmediator")
	}

	m.store = store

	ap, err := ctx.GetAriesContext()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries contect for ariesmediator")
	}

	routerClient, err := ariesmediator.New(ap)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize aries mediator client")
	}

	m.client = routerClient
	ch := make(chan service.DIDCommAction)
	m.mediatorActionChan = ch

	go m.Start()

	err = m.client.RegisterActionEvent(ch)
	if err != nil {
		return nil, errors.Wrap(err, "unable to register mediator action listener")
	}

	simp := framework.NewSimpleProvider(ap)

	bouncer, err := didexchange.NewBouncer(simp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get bouncer")
	}

	m.bouncer = bouncer

	m.vdriReg = ap.VDRIRegistry()

	return m, nil
}

func (r *Mediator) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterMediatorServer(server, r)
}

// RegisterGRPCGateway is a no-op for an internal component
func (r *Mediator) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {
}

// APISpec is a no-op for an internal component
func (r *Mediator) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *Mediator) RegisterEdgeAgent(_ context.Context, request *common.RegisterEdgeAgentRequest) (*common.RegisterEdgeAgentResponse, error) {

	if request.Secret != r.edgeAgentSecret {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid edge agent secret"))
	}

	did, err := r.store.GetMediatorDID()
	if err != nil {
		return nil, status.Error(codes.Internal, "unable to load mediator Public with DID, system must be seeded with an identity")
	}

	invite, err := r.bouncer.CreateInvitationWithDIDNotify("Canis Mediator", did.DID.String(), r.accepted(request.ExternalId), failed)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "error creating invitation with public DID for mediator")
	}

	id, err := r.store.RegisterEdgeAgent(invite.ID, request.ExternalId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to register edge agent with ID %s", request.ExternalId))

	}

	d, _ := json.MarshalIndent(invite, " ", " ")
	out := &common.RegisterEdgeAgentResponse{
		Id:         id,
		Invitation: string(d),
	}
	return out, nil
}

func (r *Mediator) GetEndpoint(_ context.Context, _ *common.EndpointRequest) (*common.EndpointResponse, error) {
	return &common.EndpointResponse{Endpoint: r.external}, nil
}

func (r *Mediator) Start() {
	for evt := range r.mediatorActionChan {
		ep := evt.Properties.All()

		theirDID, ok := ep["TheirDID"].(string)
		if !ok {
			log.Println("invalid event properties in mediator action handler")
			evt.Stop(errors.New("invalid event"))
			continue
		}

		_, err := r.store.GetEdgeAgentForDID(theirDID)
		if err != nil {
			log.Println("attempt to register for routing from unregistered DID", theirDID)
			evt.Stop(errors.New("unauthrozied registration attempts"))
			continue
		}

		//TODO: do we need to set ServiceEndpoint and Keys here?
		evt.Continue(mediator.Options{})
	}
}

func (r *Mediator) accepted(externalID string) func(id string, conn *ariesdidex.Connection) {
	return func(id string, conn *ariesdidex.Connection) {

		ea, err := r.store.GetEdgeAgent(id)
		if err != nil {
			log.Println("unable to get edge agent from conneciton", conn.ConnectionID, id)
			return
		}

		if ea.ExternalID != externalID {
			log.Printf("external id of edge agent (%s) does not match (%s) for connection %s\n", ea.ExternalID, externalID, conn.ConnectionID)
			return
		}

		ea.TheirDID = conn.TheirDID
		ea.MyDID = conn.MyDID

		err = r.store.UpdateEdgeAgent(ea)
		if err != nil {
			log.Println("error updating edge agent configuration", err)
			return
		}

		log.Printf("Successfully connected to edge agent %s to connection %s", id, "succeeded!")
	}
}

func failed(id string, err error) {
	log.Println("Connection to", id, "failed with error:", err)
}
