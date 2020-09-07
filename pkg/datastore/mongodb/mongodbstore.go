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
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/aries/storage/mongodb/store"
	"github.com/scoir/canis/pkg/datastore"
)

const (
	PublicDIDC       = "PublicDID"
	DIDC             = "DID"
	AgentC           = "Agent"
	AgentConnectionC = "AgentConnection"
	SchemaC          = "Schema"
)

type Config struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

// Provider represents a Mongo DB implementation of the storage.Provider interface
type Provider struct {
	client *mongo.Client
	dbURL  string
	stores map[string]*mongoDBStore
	dbName string
	sync.RWMutex
}

type mongoDBStore struct {
	dbURL string
	db    *mongo.Database
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
		dbURL:  config.URL,
		client: mongoClient,
		dbName: config.Database,
		stores: map[string]*mongoDBStore{}}

	return p, nil
}

// OpenStore opens and returns the collection for given name space.
func (r *Provider) OpenStore(name string) (datastore.Store, error) {
	r.Lock()
	defer r.Unlock()

	db := r.client.Database(r.dbName)

	theStore := &mongoDBStore{
		dbURL: r.dbURL,
		db:    db,
	}

	r.stores[name] = theStore

	return theStore, nil
}

func (r *mongoDBStore) GetAriesProvider() (storage.Provider, error) {
	return store.NewProvider(r.dbURL, r.db.Name()), nil
}

// Close closes the provider.
func (r *Provider) Close() error {
	r.Lock()
	defer r.Unlock()

	r.stores = make(map[string]*mongoDBStore)

	return r.client.Disconnect(context.Background())
}

// CloseStore closes a previously opened store
func (r *Provider) CloseStore(name string) error {
	r.Lock()
	defer r.Unlock()

	_, exists := r.stores[name]
	if !exists {
		return nil
	}

	delete(r.stores, name)

	return nil
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
func (r *mongoDBStore) GetSchema(id string) (*datastore.Schema, error) {
	schema := &datastore.Schema{}

	err := r.db.Collection(SchemaC).FindOne(context.Background(), bson.M{"id": id}).Decode(schema)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load schema")
	}

	return schema, nil
}

// DeleteSchema delete single schema
func (r *mongoDBStore) DeleteSchema(id string) error {
	_, err := r.db.Collection(SchemaC).DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "unable to delete schema")
	}

	return nil
}

// UpdateSchema update single schema
func (r *mongoDBStore) UpdateSchema(s *datastore.Schema) error {
	_, err := r.db.Collection(SchemaC).UpdateOne(context.Background(), bson.M{"id": s.ID}, bson.M{"$set": s})
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
func (r *mongoDBStore) GetAgent(id string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.db.Collection(AgentC).FindOne(context.Background(), bson.M{"id": id}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent")
	}

	return agent, nil

}

// GetAgentByInvitation return single agent
func (r *mongoDBStore) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.db.Collection(AgentC).FindOne(context.Background(), bson.M{"invitationid": invitationID}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent by invitation")
	}

	return agent, nil

}

// DeleteAgent delete single agent
func (r *mongoDBStore) DeleteAgent(id string) error {
	_, err := r.db.Collection(AgentC).DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "unable to delete agent")
	}

	return nil
}

// UpdateAgent delete single agent
func (r *mongoDBStore) UpdateAgent(a *datastore.Agent) error {
	_, err := r.db.Collection(AgentC).UpdateOne(context.Background(), bson.M{"id": a.ID}, bson.M{"$set": a})
	if err != nil {
		return errors.Wrap(err, "unable to update agent")
	}

	return nil
}

func (r *mongoDBStore) GetAgentConnection(a *datastore.Agent, externalID string) (*datastore.AgentConnection, error) {
	ac := &datastore.AgentConnection{}
	err := r.db.Collection(AgentConnectionC).FindOne(context.Background(),
		bson.M{"agentid": a.ID, "externalid": externalID}).Decode(ac)

	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed load agent connection").Error())
	}

	return ac, nil

}
