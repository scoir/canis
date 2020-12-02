/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	PublicDIDC           = "PublicDID"
	DIDC                 = "DID"
	AgentC               = "Agent"
	AgentConnectionC     = "AgentConnection"
	SchemaC              = "Schema"
	CredentialC          = "Credential"
	PresentationRequestC = "PresentationRequest"
	WebhookC             = "Webhook"
)

type Config struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

// Provider represents a Mongo DB implementation of the storage.Provider interface
type Provider struct {
	client *mongo.Client
	dbURL  string
	store  *mongoDBStore
	dbName string
	sync.RWMutex
}

type mongoDBStore struct {
	dbURL  string
	dbName string
	db     *mongo.Database
}

// NewProvider instantiates Provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		return nil, errors.New("config missing")
	}

	var err error
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()
	clientOpts := options.Client().SetRegistry(reg).ApplyURI(config.URL)

	mongoClient, err := mongo.NewClient(clientOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error creating mongo client")
	}

	err = mongoClient.Connect(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to mongo")
	}

	p := &Provider{
		dbName: config.Database,
		dbURL:  config.URL,
		client: mongoClient,
	}

	return p, nil
}

// Open opens and returns the collection for given name space.
func (r *Provider) Open() (datastore.Store, error) {
	r.Lock()
	defer r.Unlock()

	db := r.client.Database(r.dbName)

	theStore := &mongoDBStore{
		dbURL:  r.dbURL,
		dbName: r.dbName,
		db:     db,
	}

	r.store = theStore

	return theStore, nil
}

// Close closes the provider.
func (r *Provider) Close() error {
	r.Lock()
	defer r.Unlock()

	r.store = nil

	return r.client.Disconnect(context.Background())
}

// InsertDID add DID to store
func (r *mongoDBStore) InsertDID(d *datastore.DID) error {
	if d.DID == nil {
		return errors.New("did is required")
	}
	d.ID = d.DID.String()
	_, err := r.db.Collection(DIDC).InsertOne(context.Background(), d)
	if err != nil {
		return errors.Wrap(err, "unable to insert DID")
	}

	return nil
}

func (r *mongoDBStore) GetDID(id string) (*datastore.DID, error) {
	did := &datastore.DID{}

	err := r.db.Collection(DIDC).FindOne(context.Background(), bson.M{"id": id}).Decode(did)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load did")
	}

	return did, nil
}

// ListDIDs query DIDs
func (r *mongoDBStore) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	if c == nil {
		c = &datastore.DIDCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	bc := bson.M{}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.db.Collection(DIDC).CountDocuments(ctx, bc)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to count did docs")
	}

	results, err := r.db.Collection(DIDC).Find(ctx, bc, opts)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find DIDs")
	}

	out := datastore.DIDList{
		Count: int(count),
		DIDs:  []*datastore.DID{},
	}

	err = results.All(ctx, &out.DIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode DIDs")
	}

	return &out, nil
}

// SetPublicDID update single DID to public, unset remaining
func (r *mongoDBStore) SetPublicDID(d *datastore.DID) error {
	ctx := context.Background()
	_, err := r.db.Collection(PublicDIDC).DeleteMany(ctx, bson.M{})
	if err != nil {
		return errors.Wrap(err, "unable to unset public DID")
	}

	d.ID = d.DID.String()
	d.Public = true
	_, err = r.db.Collection(PublicDIDC).InsertOne(ctx, d)
	if err != nil {
		return errors.Wrap(err, "unable to unset public DID")
	}

	return nil
}

// GetPublicDID get public DID
func (r *mongoDBStore) GetPublicDID() (*datastore.DID, error) {
	out := &datastore.DID{}

	err := r.db.Collection(PublicDIDC).FindOne(context.Background(), bson.M{}).Decode(out)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find public PeerDID")
	}

	return out, nil
}

// InsertSchema add Schema to store
func (r *mongoDBStore) InsertSchema(s *datastore.Schema) (string, error) {
	_, err := r.db.Collection(SchemaC).InsertOne(context.Background(), s)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert schema")
	}
	return s.ID, nil
}

// ListSchema query schemas
func (r *mongoDBStore) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	bc := bson.M{}
	if c.Name != "" {
		p := fmt.Sprintf(".*%s.*", c.Name)
		bc["name"] = primitive.Regex{Pattern: p, Options: ""}
	}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()

	count, err := r.db.Collection(SchemaC).CountDocuments(ctx, bc)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to count schema")
	}

	results, err := r.db.Collection(SchemaC).Find(ctx, bc, opts)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find schema")
	}

	out := datastore.SchemaList{
		Count:  int(count),
		Schema: []*datastore.Schema{},
	}

	err = results.All(ctx, &out.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode schema")
	}

	return &out, nil
}

// GetSchema return single Schema
func (r *mongoDBStore) GetSchema(name string) (*datastore.Schema, error) {
	schema := &datastore.Schema{}

	err := r.db.Collection(SchemaC).FindOne(context.Background(), bson.M{"name": name}).Decode(schema)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load schema")
	}

	return schema, nil
}

// DeleteSchema delete single schema
func (r *mongoDBStore) DeleteSchema(name string) error {
	_, err := r.db.Collection(SchemaC).DeleteOne(context.Background(), bson.M{"name": name})
	if err != nil {
		return errors.Wrap(err, "unable to delete schema")
	}

	return nil
}

// UpdateSchema update single schema
func (r *mongoDBStore) UpdateSchema(s *datastore.Schema) error {
	_, err := r.db.Collection(SchemaC).UpdateOne(context.Background(), bson.M{"name": s.Name}, bson.M{"$set": s})
	if err != nil {
		return errors.Wrap(err, "unable to update schema")
	}

	return nil
}

// InsertAgent add agent to store
func (r *mongoDBStore) InsertAgent(a *datastore.Agent) (string, error) {
	_, err := r.db.Collection(AgentC).InsertOne(context.Background(), a)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert agent")
	}
	return a.ID, nil
}

func (r *mongoDBStore) InsertAgentConnection(a *datastore.Agent, externalID string, conn *didexchange.Connection) error {
	ac := &datastore.AgentConnection{
		AgentID:      a.ID,
		TheirLabel:   conn.TheirLabel,
		TheirDID:     conn.TheirDID,
		MyDID:        conn.MyDID,
		ConnectionID: conn.ConnectionID,
		ExternalID:   externalID,
	}

	_, err := r.db.Collection(AgentConnectionC).InsertOne(context.Background(), ac)
	if err != nil {
		return errors.Wrap(err, "unable to insert agent")
	}

	return nil

}

// ListAgent query agents
func (r *mongoDBStore) ListAgent(c *datastore.AgentCriteria) (*datastore.AgentList, error) {
	if c == nil {
		c = &datastore.AgentCriteria{
			Start:    0,
			PageSize: 10,
		}
	}

	bc := bson.M{}
	if c.Name != "" {
		p := fmt.Sprintf(".*%s.*", c.Name)
		bc["name"] = primitive.Regex{Pattern: p, Options: ""}
	}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.db.Collection(AgentC).CountDocuments(ctx, bc)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to count agents")
	}

	results, err := r.db.Collection(AgentC).Find(ctx, bc, opts)
	if err != nil {
		return nil, errors.Wrap(err, "error trying to find agents")
	}

	out := datastore.AgentList{
		Count:  int(count),
		Agents: []*datastore.Agent{},
	}

	err = results.All(ctx, &out.Agents)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode agents")
	}

	return &out, nil
}

// GetAgent return single agent
func (r *mongoDBStore) GetAgent(name string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.db.Collection(AgentC).FindOne(context.Background(), bson.M{"name": name}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent")
	}

	return agent, nil

}

// GetAgentByInvitation return single agent
func (r *mongoDBStore) GetAgentByPublicDID(DID string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	d := identifiers.ParseDID(DID)

	err := r.db.Collection(AgentC).FindOne(context.Background(), bson.M{"publicdid.did.didval.methodspecificid": d.MethodSpecificID}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent by Public DID")
	}

	return agent, nil

}

// DeleteAgent delete single agent
func (r *mongoDBStore) DeleteAgent(name string) error {
	_, err := r.db.Collection(AgentC).DeleteOne(context.Background(), bson.M{"name": name})
	if err != nil {
		return errors.Wrap(err, "unable to delete agent")
	}

	return nil
}

// UpdateAgent delete single agent
func (r *mongoDBStore) UpdateAgent(a *datastore.Agent) error {
	_, err := r.db.Collection(AgentC).UpdateOne(context.Background(), bson.M{"name": a.Name}, bson.M{"$set": a})
	if err != nil {
		return errors.Wrap(err, "unable to update agent")
	}

	return nil
}

func (r *mongoDBStore) ListAgentConnections(a *datastore.Agent) ([]*datastore.AgentConnection, error) {
	var ac []*datastore.AgentConnection
	err := r.db.Collection(AgentConnectionC).FindOne(context.Background(),
		bson.M{"agentid": a.ID}).Decode(&ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to list agent connections")
	}

	return ac, nil
}

func (r *mongoDBStore) DeleteAgentConnection(a *datastore.Agent, externalID string) error {
	_, err := r.db.Collection(AgentConnectionC).DeleteMany(context.Background(),
		bson.M{"agentid": a.ID, "externalid": externalID})

	if err != nil {
		return errors.Wrap(err, "unable to delete agent connection")
	}

	return nil
}

func (r *mongoDBStore) GetAgentConnection(a *datastore.Agent, externalID string) (*datastore.AgentConnection, error) {
	ac := &datastore.AgentConnection{}
	err := r.db.Collection(AgentConnectionC).FindOne(context.Background(),
		bson.M{"agentid": a.ID, "externalid": externalID}).Decode(ac)

	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent connection")
	}

	return ac, nil
}

func (r *mongoDBStore) InsertCredential(c *datastore.Credential) (string, error) {
	res, err := r.db.Collection(CredentialC).InsertOne(context.Background(), c)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert credential")
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), nil
}

func (r *mongoDBStore) FindOffer(offerID string) (*datastore.Credential, error) {
	c := &datastore.Credential{}
	err := r.db.Collection(CredentialC).FindOne(context.Background(),
		bson.M{"threadid": offerID, "systemstate": "offered"}).Decode(c)

	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed load offer").Error())
	}

	return c, nil
}

func (r *mongoDBStore) ListWebhooks(typ string) ([]*datastore.Webhook, error) {
	ctx := context.Background()

	results, err := r.db.Collection(WebhookC).Find(ctx, bson.M{"type": typ})

	if err != nil {
		return nil, errors.Wrap(err, "error trying to find agents")
	}

	var out []*datastore.Webhook
	err = results.All(ctx, &out)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode agents")
	}

	return out, nil
}

func (r *mongoDBStore) AddWebhook(hook *datastore.Webhook) error {
	ctx := context.Background()

	results, err := r.db.Collection(WebhookC).Find(ctx, bson.M{"type": hook.Type, "url": hook.URL})

	if err == nil && results.Next(ctx) {
		return errors.Wrapf(err, "webhook already exists for type %s", hook.Type)
	}

	_, err = r.db.Collection(WebhookC).InsertOne(ctx, hook)
	return errors.Wrap(err, "unable to insert hook")
}

func (r *mongoDBStore) DeleteWebhook(typ string) error {
	_, err := r.db.Collection(WebhookC).DeleteMany(context.Background(), bson.M{"type": typ})
	return errors.Wrap(err, "unable to remove webhook")
}

func (r *mongoDBStore) InsertPresentationRequest(pr *datastore.PresentationRequest) (string, error) {

	res, err := r.db.Collection(PresentationRequestC).InsertOne(context.Background(), pr)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *mongoDBStore) DeleteOffer(offerID string) error {
	_, err := r.db.Collection(CredentialC).DeleteOne(context.Background(), bson.M{"threadid": offerID})
	if err != nil {
		return errors.Wrap(err, "unable to delete offer")
	}

	return nil
}

func (r *mongoDBStore) GetAgentConnectionForDID(a *datastore.Agent, theirDID string) (*datastore.AgentConnection, error) {
	ac := &datastore.AgentConnection{}
	err := r.db.Collection(AgentConnectionC).FindOne(context.Background(),
		bson.M{"agentid": a.ID, "theirdid": theirDID}).Decode(ac)

	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed load agent connection").Error())
	}

	return ac, nil
}
