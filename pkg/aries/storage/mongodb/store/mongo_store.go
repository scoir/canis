/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
)

type Data struct {
	Key   string `bson:"_id" json:"Key"`
	Value []byte `bson:"Value" json:"Value"`
}

// Provider leveldb implementation of storage.Provider interface
type Provider struct {
	dial, dbname string
	lock         sync.RWMutex
	client       *mongo.Client
	db           *mongo.Database
	collections  map[string]*mongodbStore
}

// NewProvider instantiates Provider
func NewProvider(dial, db string) *Provider {
	return &Provider{dial: dial, dbname: db, collections: map[string]*mongodbStore{}}
}

// OpenStore opens and returns a store for given name space.
func (r *Provider) OpenStore(name string) (storage.Store, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	store, ok := r.collections[name]
	if ok {
		return store, nil
	}

	if r.client == nil {
		mongoClient, err := mongo.NewClient(options.Client().ApplyURI(r.dial))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create new mongo client opening store")
		}
		err = mongoClient.Connect(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "unable to connect to mongo opening a new store")
		}
		r.client = mongoClient
		r.db = r.client.Database(r.dbname)
	}

	store = &mongodbStore{
		coll: r.db.Collection(name),
	}
	r.collections[name] = store

	return store, nil
}

// Close closes all stores created under this store provider
func (r *Provider) Close() error {
	err := r.client.Disconnect(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to disconnect from mongo")
	}

	return nil
}

// CloseStore closes level dbname store of given name
func (r *Provider) CloseStore(name string) error {
	k := strings.ToLower(name)

	_, ok := r.collections[k]
	if ok {
		delete(r.collections, k)
	}

	return nil
}

type mongodbStore struct {
	coll *mongo.Collection
}

// Put stores the key and the record
func (r *mongodbStore) Put(k string, v []byte) error {
	if k == "" || v == nil {
		return errors.New("key and value are mandatory")
	}

	opts := &options.UpdateOptions{}
	_, err := r.coll.UpdateOne(context.Background(), bson.M{"_id": k}, bson.M{"$set": Data{Key: k, Value: v}},
		opts.SetUpsert(true))

	return err
}

// Get fetches the record based on key
func (r *mongodbStore) Get(k string) ([]byte, error) {
	if k == "" {
		return nil, errors.New("key is mandatory")
	}

	data := &Data{}
	result := r.coll.FindOne(context.Background(), bson.M{"_id": k})
	if result.Err() == mongo.ErrNoDocuments {
		return nil, storage.ErrDataNotFound
	} else if result.Err() != nil {
		return nil, errors.Wrap(result.Err(), "unable to query mongo")
	}

	err := result.Decode(data)
	if err != nil {
		return nil, errors.Wrap(err, "invalid data storage, mongo store")
	}

	return data.Value, nil
}

// Iterator returns iterator for the latest snapshot of the underlying dbname.
func (r *mongodbStore) Iterator(start, limit string) storage.StoreIterator {
	if start == "" || limit == "" {
		return iterator.NewEmptyIterator(errors.New("start or limit key is mandatory"))
	}

	cur, err := r.coll.Find(context.Background(), bson.M{})
	if err != nil {
		return nil
	}

	return &mongodbIterator{cursor: cur}

}

// Delete will delete record with k key
func (r *mongodbStore) Delete(k string) error {
	if k == "" {
		return errors.New("key is mandatory")
	}

	_, err := r.coll.DeleteOne(context.Background(), bson.M{"_id": k})
	return err
}

type mongodbIterator struct {
	cursor *mongo.Cursor
}

func (r *mongodbIterator) Next() bool {
	return r.cursor.Next(context.Background())
}

func (r *mongodbIterator) Release() {
	_ = r.cursor.Close(context.Background())
}

func (r *mongodbIterator) Error() error {
	return r.cursor.Err()
}

func (r *mongodbIterator) Key() []byte {
	d := &Data{}
	err := r.cursor.Decode(d)
	if err != nil {
		return nil
	}
	return []byte(d.Key)
}

func (r *mongodbIterator) Value() []byte {
	d := &Data{}
	err := r.cursor.Decode(d)
	if err != nil {
		return nil
	}
	return d.Value
}
