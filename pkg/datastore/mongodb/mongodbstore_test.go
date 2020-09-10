/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package mongodb

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
)

const (
	mongoStoreDBURL = "mongodb://localhost:27017"
)

var (
	config = &Config{
		URL:      mongoStoreDBURL,
		Database: "test",
	}
)

//docker run --name some-mongo -d mongo:tag
// For these unit tests to run, you must ensure you have a Mongo DB instance running at the URL specified in
// sqlStoreDBURL.
// To run the tests manually, start an instance by running the following command in the terminal
// docker run -p 27017:27017 --name MongoStoreTest -d mongo:4.2.8
// delete using
//   docker kill MongoStoreTest
//   docker rm MongoStoreTest
func TestMain(m *testing.M) {
	err := waitForMongoDBToStart()
	if err != nil {
		fmt.Printf(err.Error() +
			". Make sure you start a mongoDBStore instance using" +
			" 'docker run -p 5432:5432 mongo:4.2.8' before running the unit tests")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func waitForMongoDBToStart() error {
	var err error
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()
	clientOpts := options.Client().SetRegistry(reg).ApplyURI(mongoStoreDBURL)

	mongoClient, err := mongo.NewClient(clientOpts)
	if err != nil {
		return err
	}

	err = mongoClient.Connect(context.Background())
	if err != nil {
		return errors.Wrap(err, "error connecting to mongo")
	}
	db := mongoClient.Database("test")

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: couldn't reach no sql db server")
		default:
			err := db.Client().Ping(context.Background(), nil)
			if err != nil {
				return err
			}

			return nil
		}
	}
}

func TestInsertListDID(t *testing.T) {
	t.Run("Test insert / list public did", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_list")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "a did",
				},
			},
			Public: true,
		})
		require.NoError(t, err)

		didlist, err := store.ListDIDs(&datastore.DIDCriteria{})
		require.NoError(t, err)

		require.Equal(t, didlist.Count, 1)

		did := didlist.DIDs[0].DID.DIDVal.MethodSpecificID
		require.Equal(t, "a did", did)

		err = prov.CloseStore("test_list")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestSetGetPublicDID(t *testing.T) {
	t.Run("Test get / set public did", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_dids")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "did to be public",
			},
		}, Public: false})
		require.NoError(t, err)
		d := &datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "didtobepublic",
			},
		}, Public: true}

		err = store.SetPublicDID(d)
		require.NoError(t, err)

		public, err := store.GetPublicDID()
		require.NoError(t, err)

		require.Equal(t, "didtobepublic", public.DID.DIDVal.MethodSpecificID)

		err = prov.CloseStore("test_dids")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestSchema(t *testing.T) {
	t.Run("Test schema CRUD", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_schemas")
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{ID: "schema id", Name: "schema name"})
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{ID: "another schema id", Name: "another schema name"})
		require.NoError(t, err)

		list, err := store.ListSchema(&datastore.SchemaCriteria{})
		require.NoError(t, err)
		require.Equal(t, 2, list.Count)

		err = store.UpdateSchema(&datastore.Schema{ID: "schema id", Name: "different schema name"})
		require.NoError(t, err)

		updated, err := store.GetSchema("schema id")
		require.NoError(t, err)
		require.Equal(t, "different schema name", updated.Name)

		err = store.DeleteSchema("schema id")
		require.NoError(t, err)

		list, err = store.ListSchema(&datastore.SchemaCriteria{Name: "another"})
		require.NoError(t, err)
		require.Equal(t, 1, list.Count)

		err = prov.CloseStore("test_schemas")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestAgent(t *testing.T) {
	t.Run("Test Agent CRUD", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_agents")
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{ID: "agent id", Name: "an agent"})
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{ID: "agent id 2", Name: "another agent"})
		require.NoError(t, err)

		list, err := store.ListAgent(nil)
		require.NoError(t, err)
		require.Equal(t, 2, list.Count)

		err = store.UpdateAgent(&datastore.Agent{ID: "agent id", Name: "an different agent"})
		require.NoError(t, err)

		updated, err := store.GetAgent("agent id")
		require.NoError(t, err)
		require.Equal(t, "an different agent", updated.Name)

		err = store.DeleteAgent("agent id")
		require.NoError(t, err)

		list, err = store.ListAgent(&datastore.AgentCriteria{Name: "anoth"})
		require.NoError(t, err)
		require.Equal(t, 1, list.Count)

		err = prov.CloseStore("test_agents")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestProviderFailures(t *testing.T) {
	t.Run("no config error", func(t *testing.T) {
		_, err := NewProvider(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "config missing")
	})

	t.Run("config params error", func(t *testing.T) {
		_, err := NewProvider(&Config{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "error creating mongo client")
	})
}

func TestCloseStore(t *testing.T) {
	t.Run("no config error", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		_, err = prov.OpenStore("test_closing")
		require.NoError(t, err)

		err = prov.CloseStore("test_closing")
		require.NoError(t, err)
	})

	t.Run("no mapping found", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		_, err = prov.OpenStore("test_closing")
		require.NoError(t, err)

		delete(prov.stores, "test_closing")

		err = prov.CloseStore("test_closing")
		require.NoError(t, err)
	})
}

func TestDIDFailures(t *testing.T) {
	t.Run("connectivity failures", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_did_failures")
		require.NoError(t, err)

		err = prov.client.Disconnect(context.Background())
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "did is required")

		_, err = store.ListDIDs(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error trying to find DIDs")

		_, err = store.GetPublicDID()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to find public")
	})
}

func TestSchemaFailures(t *testing.T) {
	t.Run("connectivity failures", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_schema_failures")
		require.NoError(t, err)

		err = prov.client.Disconnect(context.Background())
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to insert schema")

		err = store.DeleteSchema("foo")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to delete schema")

		err = store.UpdateSchema(&datastore.Schema{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to update schema")

		_, err = store.GetSchema("bar")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to load schema")

		_, err = store.ListSchema(&datastore.SchemaCriteria{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "error trying to find schema")
	})
}

func TestAgentFailures(t *testing.T) {
	t.Run("connectivity failures", func(t *testing.T) {
		prov, err := NewProvider(config)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_agent_failures")
		require.NoError(t, err)

		err = prov.client.Disconnect(context.Background())
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{ID: "agent id", Name: "an agent"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to insert agent")

		_, err = store.GetAgentByPublicDID("foo")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to load agent by Public DID")

		err = store.DeleteAgent("foo")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to delete agent")

		err = store.UpdateAgent(&datastore.Agent{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to update agent")

		_, err = store.GetAgent("bar")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to load agent")

		_, err = store.ListAgent(&datastore.AgentCriteria{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "error trying to find agents")
	})
}
