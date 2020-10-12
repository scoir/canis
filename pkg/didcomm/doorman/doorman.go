package doorman

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/btcsuite/btcutil/base58"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	ariesdidex "github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/doorman/api"
	"github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
	"github.com/scoir/canis/pkg/notifier"
)

const ConnectionTopic = "connections"
const AcceptedEvent = "accepted"

type provider interface {
	GetAriesContext() (*ariescontext.Provider, error)
	GetDatastore() (datastore.Store, error)
	GetAMQPConfig() *framework.AMQPConfig
}

type Doorman struct {
	store               datastore.Store
	didcl               *ariesdidex.Client
	bouncer             didexchange.Bouncer
	vdriReg             vdriapi.Registry
	notificationChannel *amqp.Channel
}

func New(prov provider) (*Doorman, error) {

	ctx, err := prov.GetAriesContext()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load aries context")
	}

	simp := framework.NewSimpleProvider(ctx)
	bouncer, err := didexchange.NewBouncer(simp)
	vdriReg := ctx.VDRIRegistry()

	agentStore, err := prov.GetDatastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get datastore provider")
	}

	conn, err := amqp.Dial(prov.GetAMQPConfig().Endpoint())
	if err != nil {
		log.Fatalln("unable to dial RMQ", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create an AMQP channel")
	}

	_, err = ch.QueueDeclare(
		notifier.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to declare AMQP queue")
	}

	return &Doorman{
		store:               agentStore,
		bouncer:             bouncer,
		vdriReg:             vdriReg,
		notificationChannel: ch,
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

	agent, err := r.store.GetAgent(request.AgentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("agent with id %s not found", request.AgentId))
	}

	_, err = r.store.GetAgentConnection(agent, request.ExternalId)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists,
			fmt.Sprintf("connection between agent %s and external ID %s already exists", agent.ID, request.ExternalId))
	}

	var invite *ariesdidex.Invitation
	if agent.HasPublicDID {
		did := agent.PublicDID.DID.String()
		invite, err = r.bouncer.CreateInvitationWithDIDNotify(agent.Name, did, r.accepted(agent, request.ExternalId), failed)
		if err != nil {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("schema with id %s already exists", request.AgentId))
		}
	} else {
		invite, err = r.bouncer.CreateInvitationNotify(agent.Name, r.accepted(agent, request.ExternalId), failed)
		if err != nil {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("schema with id %s already exists", request.AgentId))
		}
	}

	//invite.Type = "did:sov:BzCbsNYhMrjHiqZDTUASHg;spec/connections/1.0/invitation"
	d, _ := json.MarshalIndent(invite, " ", " ")
	return &api.InvitationResponse{
		Invitation: string(d),
	}, nil
}

func (r *Doorman) accepted(agent *datastore.Agent, externalID string) func(id string, conn *ariesdidex.Connection) {
	return func(id string, conn *ariesdidex.Connection) {
		err := r.store.InsertAgentConnection(agent, externalID, conn)
		if err != nil {
			log.Println("error creating agent connection", err)
			return
		}

		//GET PEER DID FROM REGISTRY, SAVE.
		diddoc, err := r.vdriReg.Resolve(conn.MyDID)
		if err != nil {
			log.Println("error resolving peer DID", err)
			return
		}

		pubKey := base58.Encode(diddoc.PublicKey[0].Value)
		did := &datastore.DID{
			ID: diddoc.ID,
			DID: &identifiers.DID{
				DIDVal: identifiers.ParseDID(diddoc.ID),
				Verkey: pubKey,
			},
			OwnerID: agent.ID,
			KeyPair: &datastore.KeyPair{
				ID:        diddoc.PublicKey[0].ID[1:],
				PublicKey: pubKey,
			},
			Endpoint: diddoc.Service[0].ServiceEndpoint,
			Public:   false,
		}
		err = r.store.InsertDID(did)
		if err != nil {
			log.Println("error saving peer DID", err)
			return
		}

		r.publishEvent(agent, externalID, conn)
		log.Printf("Successfully connected agent %s to connection %s", id, "succeeded!")
	}
}

func (r *Doorman) publishEvent(agent *datastore.Agent, externalID string, conn *ariesdidex.Connection) {

	evt := &notifier.Notification{
		Topic: ConnectionTopic,
		Event: AcceptedEvent,
		EventData: DIDAcceptedEvent{
			AgentID:      agent.ID,
			TheirDID:     conn.TheirDID,
			MyDID:        conn.MyDID,
			ConnectionID: conn.ConnectionID,
			ExternalID:   externalID,
		},
	}

	message, err := json.Marshal(evt)
	if err != nil {
		log.Printf("unexpected error marshalling did accepted event")
		return
	}

	err = r.notificationChannel.Publish(
		"",
		notifier.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})

	if err != nil {
		log.Printf("unable to publish doorman event")
		return
	}
}

func failed(id string, err error) {
	log.Println("Connection to", id, "failed with error:", err)
}
