/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mongodb

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/scoir/canis/pkg/datastore"
)

var (
	SchemaC = "Schema"
	AgentC  = "Agent"
	DIDC    = "PeerDID"
)

type Store struct {
	database *mongo.Database
}

func (r *Store) InsertDID(d *datastore.DID) error {
	_, err := r.database.Collection(DIDC).InsertOne(context.Background(), d)
	if err != nil {
		return errors.Wrap(err, "unable to insert PeerDID")
	}

	return nil
}

func (r *Store) ListDIDs(c *datastore.DIDCriteria) (*datastore.DIDList, error) {
	bc := bson.M{}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.database.Collection(DIDC).CountDocuments(ctx, bc)
	results, err := r.database.Collection(DIDC).Find(ctx, bc, opts)

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

func (r *Store) SetPublicDID(DID string) error {
	ctx := context.Background()
	_, err := r.database.Collection(DIDC).UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"public": false}})
	if err != nil {
		return errors.Wrap(err, "unable to unset public PeerDID")
	}

	_, err = r.database.Collection(DIDC).UpdateOne(ctx, bson.M{"did": DID}, bson.M{"$set": bson.M{"public": true}})
	if err != nil {
		return errors.Wrap(err, "unable to unset public PeerDID")
	}

	return nil
}

func (r *Store) GetPublicDID() (*datastore.DID, error) {
	out := &datastore.DID{}
	err := r.database.Collection(DIDC).FindOne(context.Background(), bson.M{"public": true}).Decode(out)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find public PeerDID")
	}

	return out, nil
}

func NewStore(db *mongo.Database) *Store {
	return &Store{database: db}
}

func (r *Store) InsertSchema(s *datastore.Schema) (string, error) {
	_, err := r.database.Collection(SchemaC).InsertOne(context.Background(), s)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert schema")
	}
	return s.ID, nil
}

func (r *Store) ListSchema(c *datastore.SchemaCriteria) (*datastore.SchemaList, error) {
	bc := bson.M{}
	if c.Name != "" {
		p := fmt.Sprintf(".*%s.*", c.Name)
		bc["name"] = primitive.Regex{Pattern: p, Options: ""}
	}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.database.Collection(SchemaC).CountDocuments(ctx, bc)
	results, err := r.database.Collection(SchemaC).Find(ctx, bc, opts)

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

func (r *Store) GetSchema(id string) (*datastore.Schema, error) {
	schema := &datastore.Schema{}

	err := r.database.Collection(SchemaC).FindOne(context.Background(), bson.M{"id": id}).Decode(schema)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load schema")
	}

	return schema, nil
}

func (r *Store) DeleteSchema(id string) error {
	_, err := r.database.Collection(SchemaC).DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "unable to delete schema")
	}

	return nil
}

func (r *Store) UpdateSchema(s *datastore.Schema) error {
	_, err := r.database.Collection(SchemaC).UpdateOne(context.Background(), bson.M{"id": s.ID}, bson.M{"$set": s})
	if err != nil {
		return errors.Wrap(err, "unable to update schema")
	}

	return nil
}

func (r *Store) InsertAgent(a *datastore.Agent) (string, error) {
	_, err := r.database.Collection(AgentC).InsertOne(context.Background(), a)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert agent")
	}
	return a.ID, nil

}

func (r *Store) ListAgent(c *datastore.AgentCriteria) (*datastore.AgentList, error) {
	bc := bson.M{}
	if c.Name != "" {
		p := fmt.Sprintf(".*%s.*", c.Name)
		bc["name"] = primitive.Regex{Pattern: p, Options: ""}
	}

	opts := &options.FindOptions{}
	opts = opts.SetSkip(int64(c.Start)).SetLimit(int64(c.PageSize))

	ctx := context.Background()
	count, err := r.database.Collection(AgentC).CountDocuments(ctx, bc)
	results, err := r.database.Collection(AgentC).Find(ctx, bc, opts)

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

func (r *Store) GetAgent(id string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.database.Collection(AgentC).FindOne(context.Background(), bson.M{"id": id}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent")
	}

	return agent, nil

}

func (r *Store) GetAgentByInvitation(invitationID string) (*datastore.Agent, error) {
	agent := &datastore.Agent{}

	err := r.database.Collection(AgentC).FindOne(context.Background(), bson.M{"InvitationID": invitationID}).Decode(agent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load agent by invitation")
	}

	return agent, nil

}

func (r *Store) DeleteAgent(id string) error {
	_, err := r.database.Collection(AgentC).DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return errors.Wrap(err, "unable to delete agent")
	}

	return nil
}

func (r *Store) UpdateAgent(a *datastore.Agent) error {
	_, err := r.database.Collection(AgentC).UpdateOne(context.Background(), bson.M{"id": a.ID}, bson.M{"$set": a})
	if err != nil {
		return errors.Wrap(err, "unable to update agent")
	}

	return nil
}
