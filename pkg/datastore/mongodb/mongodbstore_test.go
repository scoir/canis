/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	mongoStoreDBURL = "mongodb://${MONGODB_HOST}:27017/"
)

func testConfig() *Config {
	return &Config{
		URL:      os.ExpandEnv(mongoStoreDBURL),
		Database: uuid.New().String(),
	}
}

//docker run --name some-mongo -d mongo:tag
// For these unit tests to run, you must ensure you have a Mongo DB instance running at the URL specified in
// mongoStoreDBURL.
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

	res := m.Run()

	os.Exit(res)
}

func dropTestDatabase(dbName string) {
	var err error
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()
	clientOpts := options.Client().SetRegistry(reg).ApplyURI(os.ExpandEnv(mongoStoreDBURL))

	mongoClient, err := mongo.NewClient(clientOpts)
	if err != nil {
		log.Fatalln("error dropping database", err)
	}

	ctx := context.Background()
	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatalln("error dropping database", err)
	}
	db := mongoClient.Database(dbName)
	err = db.Drop(ctx)
	if err != nil {
		log.Fatalln("error dropping database", err)
	}
}

func waitForMongoDBToStart() error {
	var err error
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()
	clientOpts := options.Client().SetRegistry(reg).ApplyURI(os.ExpandEnv(mongoStoreDBURL))

	mongoClient, err := mongo.NewClient(clientOpts)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = mongoClient.Connect(ctx)
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
			err = db.Drop(ctx)
			if err != nil {
				log.Fatalln("error dropping database", err)
			}

			return nil
		}
	}
}

func TestInsertListDID(t *testing.T) {
	t.Run("Test insert / list public did", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
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

		d2, err := store.GetDID(didlist.DIDs[0].DID.String())
		require.NoError(t, err)
		require.Equal(t, "did:a did", d2.ID)

		err = prov.Close()
		require.NoError(t, err)

	})
}

func TestSetGetPublicDID(t *testing.T) {
	t.Run("Test get / set public did", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
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

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestSchema(t *testing.T) {
	t.Run("Test schema CRUD", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
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

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestAgent(t *testing.T) {
	t.Run("Test Agent CRUD", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
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

		did := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123",
				},
			},
		}
		agent := &datastore.Agent{ID: "agent id 2", Name: "another agent", PublicDID: did}
		_, err = store.InsertAgent(agent)
		require.NoError(t, err)

		agent, err = store.GetAgentByPublicDID("did:sov:123")
		require.NoError(t, err)
		require.NotNil(t, agent)

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

func TestDIDFailures(t *testing.T) {
	t.Run("connectivity failures", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
		require.NoError(t, err)

		err = prov.client.Disconnect(context.Background())
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "did is required")

		err = store.InsertDID(&datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "did to be public",
			},
		}, Public: false})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to insert DID")

		_, err = store.ListDIDs(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error trying to count did docs")

		_, err = store.GetDID("did:sov:123")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to load did")

		_, err = store.GetPublicDID()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to find public")
	})
}

func TestSchemaFailures(t *testing.T) {
	t.Run("connectivity failures", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
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
		require.Contains(t, err.Error(), "error trying to count schema")
	})
}

func TestAgentFailures(t *testing.T) {
	t.Run("connectivity failures", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
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
		require.Contains(t, err.Error(), "error trying to count agents")

		err = store.InsertAgentConnection(&datastore.Agent{}, "subject", &didexchange.Connection{
			Record: &connection.Record{},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to insert agent")

		_, err = store.GetAgentConnection(&datastore.Agent{}, "subject")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to load agent connection")
	})
}

func TestPresentationRequest(t *testing.T) {
	t.Run("insert presentation request", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
		require.NoError(t, err)

		_, err = store.InsertPresentationRequest(&datastore.PresentationRequest{
			AgentID:               "agent id",
			SchemaID:              "schema id",
			ExternalID:            "external id",
			PresentationRequestID: "presentation request id",
		})
		require.NoError(t, err)

	})
}

func TestOffer(t *testing.T) {
	t.Run("credential offer", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
		require.NoError(t, err)

		_, err = store.InsertCredential(&datastore.IssuedCredential{ThreadID: "1234", SystemState: "offered"})
		require.NoError(t, err)

		c, err := store.FindCredentialByOffer("1234")
		require.NoError(t, err)
		require.NotNil(t, c)

		err = store.DeleteCredentialByOffer("1234")
		require.NoError(t, err)

		c, err = store.FindCredentialByOffer("1234")
		require.Error(t, err)
		require.Nil(t, c)

	})
}

func TestAgentConnection(t *testing.T) {
	t.Run("agent connection", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
		require.NoError(t, err)

		agent := &datastore.Agent{ID: "agent id", Name: "an agent"}
		conn := &didexchange.Connection{
			Record: &connection.Record{
				TheirDID:     "did:sov:abc",
				MyDID:        "did:sov:xyz",
				ConnectionID: "connection-id",
			},
		}

		err = store.InsertAgentConnection(agent, "external-id", conn)
		require.NoError(t, err)

		ac, err := store.GetAgentConnection(agent, "external-id")
		require.NoError(t, err)
		require.Equal(t, ac.TheirDID, "did:sov:abc")

		ac, err = store.GetAgentConnectionForDID(agent, "did:sov:abc")
		require.NoError(t, err)
		require.Equal(t, ac.ExternalID, "external-id")
		require.Equal(t, ac.AgentName, "agent id")
		require.Equal(t, ac.MyDID, "did:sov:xyz")
		require.Equal(t, ac.ConnectionID, "connection-id")

	})
}

func TestWebhooks(t *testing.T) {
	t.Run("webhooks", func(t *testing.T) {
		conf := testConfig()
		prov, err := NewProvider(conf)
		defer dropTestDatabase(conf.Database)
		require.NoError(t, err)

		store, err := prov.Open()
		require.NoError(t, err)

		err = store.AddWebhook(&datastore.Webhook{
			Type: "connections",
			URL:  "http://example.com/connections",
		})
		require.NoError(t, err)

		err = store.AddWebhook(&datastore.Webhook{
			Type: "connections",
			URL:  "http://sample.com/connections",
		})
		require.NoError(t, err)

		err = store.AddWebhook(&datastore.Webhook{
			Type: "credentials",
			URL:  "http://sample.com/credentials",
		})
		require.NoError(t, err)

		hooks, err := store.ListWebhooks("connections")
		require.NoError(t, err)
		require.Len(t, hooks, 2)

		err = store.DeleteWebhook("connections")
		require.NoError(t, err)

		hooks, err = store.ListWebhooks("connections")
		require.NoError(t, err)
		require.Len(t, hooks, 0)
	})
}
