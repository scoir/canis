/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mongodb"
)

type DatastoreConfig struct {
	Database string `mapstructure:"database"`
	Mongo    *Mongo `mapstructure:"mongo"`

	lock sync.Mutex
	ds   datastore.Store
}

type Mongo struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

func (r *DatastoreConfig) Datastore() (datastore.Store, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.ds != nil {
		return r.ds, nil
	}

	var err error
	switch r.Database {
	case "mongo":
		r.ds, err = r.loadMongo()
	case "postgres":
		r.ds, err = r.loadPostgres()
	}

	return r.ds, errors.Wrap(err, "unable to get datastore from config")
}

func (r *DatastoreConfig) loadMongo() (datastore.Store, error) {
	mongoClient, err := getClient(r)
	if err != nil {
		return nil, err
	}

	return mongodb.NewStore(mongoClient.Database(r.Mongo.Database)), nil
}

func (r *DatastoreConfig) loadPostgres() (datastore.Store, error) {
	return nil, errors.New("not implemented")
}

func getClient(conf *DatastoreConfig) (*mongo.Client, error) {
	var err error
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()
	clientOpts := options.Client().SetRegistry(reg).ApplyURI(conf.Mongo.URL)

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
