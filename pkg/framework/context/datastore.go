/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mongodb"
	"github.com/scoir/canis/pkg/datastore/postgres"
)

const (
	datastoreKey = "datastore"
)

type DatastoreConfig struct {
	Database string           `mapstructure:"database"`
	Mongo    *Mongo           `mapstructure:"mongo"`
	Postgres *postgres.Config `mapstructure:"postgres"`
}

type Mongo struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

func (r *Provider) Datastore() (datastore.Provider, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.ds != nil {
		return r.ds, nil
	}

	dc := &DatastoreConfig{}
	err := r.vp.UnmarshalKey(datastoreKey, dc)
	if err != nil {
		return nil, errors.Wrap(err, "execution environment is not correctly configured")
	}

	switch dc.Database {
	case "mongo":
		r.ds, err = r.loadMongo(dc.Mongo)
	case "postgres":
		r.ds, err = postgres.NewProvider(dc.Postgres)
	default:
		return nil, errors.New("no datastore configuration was provided")
	}

	return r.ds, errors.Wrap(err, "unable to get datastore from config")
}

func (r *Provider) loadMongo(dsc *Mongo) (datastore.Store, error) {
	if dsc == nil {
		return nil, errors.New("mongo driver not property configured")
	}

	mongoClient, err := getClient(dsc)
	if err != nil {
		return nil, err
	}

	return mongodb.NewStore(mongoClient.Database(dsc.Database)), nil
}

func getClient(conf *Mongo) (*mongo.Client, error) {
	var err error
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()
	clientOpts := options.Client().SetRegistry(reg).ApplyURI(conf.URL)

	mongoClient, err := mongo.NewClient(clientOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error creating mongo client")
	}
	err = mongoClient.Connect(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to mongo")
	}

	return mongoClient, err
}
