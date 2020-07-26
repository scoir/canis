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

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/scoir/canis/pkg/datastore"
)

type Config struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

// Provider represents a Mongo DB implementation of the storage.Provider interface
type Provider struct {
	db       *mongo.Database
	dbURL    string
	stores   map[string]*mongoDBStore
	dbPrefix string
	sync.RWMutex
}

type mongoDBStore struct {
	collection *mongo.Collection
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
	db := mongoClient.Database(config.Database)

	p := &Provider{
		db:     db,
		stores: map[string]*mongoDBStore{}}

	return p, nil
}

// OpenStore opens and returns the collection for given name space.
func (p *Provider) OpenStore(name string) (datastore.Store, error) {
	p.Lock()
	defer p.Unlock()

	if name == "" {
		return nil, errors.New("store name is required")
	}

	store := &mongoDBStore{
		collection: p.db.Collection(name),
	}

	p.stores[name] = store

	return store, nil
}

// Close closes the provider.
func (p *Provider) Close() error {
	p.Lock()
	defer p.Unlock()

	p.stores = make(map[string]*mongoDBStore)

	return nil
}

// CloseStore closes a previously opened stores
func (p *Provider) CloseStore(name string) error {
	p.Lock()
	defer p.Unlock()

	_, exists := p.stores[name]
	if !exists {
		return nil
	}

	delete(p.stores, name)

	return p.db.Client().Disconnect(context.Background())
}

func (r *mongoDBStore) InsertDID(d *datastore.DID) error {
	_, err := r.collection.InsertOne(context.Background(), d)
	if err != nil {
		return errors.Wrap(err, "unable to insert DID")
	}

	return nil
}

func (r *mongoDBStore) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	//todo or fail if no criteria?
	c = &datastore.DIDCriteria{
		Start:    0,
		PageSize: 10,
	}

	bc := bson.M{}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.collection.CountDocuments(ctx, bc)
	results, err := r.collection.Find(ctx, bc, opts)

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

func (r *mongoDBStore) SetPublicDID(DID string) error {
	ctx := context.Background()
	_, err := r.collection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"public": false}})
	if err != nil {
		return errors.Wrap(err, "unable to unset public PeerDID")
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"did": DID}, bson.M{"$set": bson.M{"public": true}})
	if err != nil {
		return errors.Wrap(err, "unable to unset public PeerDID")
	}

	return nil
}

func (r *mongoDBStore) GetPublicDID() (*datastore.DID, error) {
	out := &datastore.DID{}
	err := r.collection.FindOne(context.Background(), bson.M{"public": true}).Decode(out)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find public PeerDID")
	}

	return out, nil
}

func (r *mongoDBStore) InsertSchema(s *datastore.Schema) (string, error) {
	_, err := r.collection.InsertOne(context.Background(), s)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert schema")
	}
	return s.ID, nil
}

func (r *mongoDBStore) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	bc := bson.M{}
	if c.Name != "" {
		p := fmt.Sprintf(".*%s.*", c.Name)
		bc["name"] = primitive.Regex{Pattern: p, Options: ""}
	}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.collection.CountDocuments(ctx, bc)
	results, err := r.collection.Find(ctx, bc, opts)

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

func (r *mongoDBStore) GetSchema(id string) (*datastore.Schema, error) {
	schema := &datastore.Schema{}

	err := r.collection.FindOne(context.Background(), bson.M{"id": id}).Decode(schema)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load schema")
	}

	return schema, nil
}

func (r *mongoDBStore) DeleteSchema(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "unable to delete schema")
	}

	return nil
}

func (r *mongoDBStore) UpdateSchema(s *datastore.Schema) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"id": s.ID}, bson.M{"$set": s})
	if err != nil {
		return errors.Wrap(err, "unable to update schema")
	}

	return nil
}

func (r *mongoDBStore) InsertAgent(a *datastore.Agent) (string, error) {
	_, err := r.collection.InsertOne(context.Background(), a)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert agent")
	}
	return a.ID, nil

}

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
	count, err := r.collection.CountDocuments(ctx, bc)
	results, err := r.collection.Find(ctx, bc, opts)

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

func (r *mongoDBStore) GetAgent(id string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.collection.FindOne(context.Background(), bson.M{"id": id}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent")
	}

	return agent, nil

}

func (r *mongoDBStore) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.collection.FindOne(context.Background(), bson.M{"invitationid": invitationID}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent by invitation")
	}

	return agent, nil

}

func (r *mongoDBStore) DeleteAgent(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "unable to delete agent")
	}

	return nil
}

func (r *mongoDBStore) UpdateAgent(a *datastore.Agent) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"id": a.ID}, bson.M{"$set": a})
	if err != nil {
		return errors.Wrap(err, "unable to update agent")
	}

	return nil
}
