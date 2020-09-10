/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdbstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kivik/kivik"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
)

const (
	couchDBURL = "localhost:5984"
)

// Useful during testing http://127.0.0.1:5984/_utils/#/_all_dbs
func TestMain(m *testing.M) {
	err := waitForCouchDBToStart()
	if err != nil {
		fmt.Printf(err.Error() +
			". Make sure you start a couchDB instance using" +
			" 'docker run -p 5984:5984 couchdb:2.3.1' before running the unit tests")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func waitForCouchDBToStart() error {
	client, err := kivik.New("couch", couchDBURL)
	if err != nil {
		return err
	}

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: couldn't reach CouchDB server")
		default:
			dbs, err := client.AllDBs(context.Background())
			if err != nil {
				return err
			}

			for _, v := range dbs {
				if err := client.DestroyDB(context.Background(), v); err != nil {
					panic(err.Error())
				}
			}

			return nil
		}
	}
}

func TestInsertListDID(t *testing.T) {
	t.Run("Test insert / list public did", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
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

		didlist, err := store.ListDIDs(nil)
		require.NoError(t, err)

		require.Equal(t, didlist.Count, 1)

		did := didlist.DIDs[0].DID.DIDVal.MethodSpecificID
		require.Equal(t, "a did", did)

		err = prov.CloseStore("test_list")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})

	t.Run("paging", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_list_paging")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "a did 1",
			},
		}})
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "a did 2",
			},
		}})
		require.NoError(t, err)

		didlist, err := store.ListDIDs(&datastore.DIDCriteria{
			Start:    1,
			PageSize: 1,
		})
		require.NoError(t, err)

		require.Equal(t, didlist.Count, 2)

		err = prov.CloseStore("test_list_paging")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestDIDErrors(t *testing.T) {
	t.Run("no did", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_did_errors")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{
			Public: true,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "malformed DID")

		err = prov.CloseStore("test_did_errors")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})

}

func TestSetGetPublicDID(t *testing.T) {
	t.Run("Test get / set public did", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_getsetdids")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "did to be public",
			},
		}, Public: false})
		require.NoError(t, err)

		err = store.SetPublicDID(&datastore.DID{DID: &identifiers.DID{
			DIDVal: identifiers.DIDValue{
				MethodSpecificID: "another did",
			},
		}, Public: true})
		require.NoError(t, err)

		public, err := store.GetPublicDID()
		require.NoError(t, err)

		require.Equal(t, "another did", public.DID.DIDVal.MethodSpecificID)

		err = prov.CloseStore("test_getsetdids")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestSchema(t *testing.T) {
	t.Run("Test schema CRUD", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_schemas")
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{ID: "schema id", Name: "schema name"})
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{ID: "another schema id", Name: "another schema name"})
		require.NoError(t, err)

		list, err := store.ListSchema(nil)
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

	t.Run("generate ID", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_schemas_ids")
		require.NoError(t, err)

		id, err := store.InsertSchema(&datastore.Schema{})
		require.NoError(t, err)
		require.NotEmpty(t, id)
	})

	t.Run("paging", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_schema_paging")
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{Name: "a schema 1"})
		require.NoError(t, err)

		_, err = store.InsertSchema(&datastore.Schema{Name: "a schema 2"})
		require.NoError(t, err)

		schemas, err := store.ListSchema(&datastore.SchemaCriteria{
			Start:    1,
			PageSize: 1,
		})
		require.NoError(t, err)

		require.Equal(t, schemas.Count, 2)

		err = prov.CloseStore("test_schema_paging")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestAgent(t *testing.T) {
	t.Run("Test Agent CRUD", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
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

	t.Run("generate ID", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_schemas_ids")
		require.NoError(t, err)

		id, err := store.InsertAgent(&datastore.Agent{})
		require.NoError(t, err)
		require.NotEmpty(t, id)
	})

	t.Run("paging", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_agent_paging")
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{Name: "agent 1"})
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{Name: "agent 2"})
		require.NoError(t, err)

		agents, err := store.ListSchema(&datastore.SchemaCriteria{
			Start:    1,
			PageSize: 1,
		})
		require.NoError(t, err)

		require.Equal(t, agents.Count, 2)

		err = prov.CloseStore("test_schema_paging")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestProviderFailures(t *testing.T) {
	t.Run("blank url error", func(t *testing.T) {
		_, err := NewProvider("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "can't be blank")
	})

	t.Run("wrong url error", func(t *testing.T) {
		prov, err := NewProvider("thisisbad")
		require.NoError(t, err)

		store, err := prov.OpenStore("sample")
		require.Error(t, err)
		require.Nil(t, store)
	})
}

func TestCloseProvider(t *testing.T) {
	t.Run("should close stores", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		_, err = prov.OpenStore("test_closing_provider")
		require.NoError(t, err)

		require.Equal(t, 1, len(prov.dbs))

		err = prov.Close()
		require.NoError(t, err)

		require.Equal(t, 0, len(prov.dbs))
	})

	t.Run("no mapping found", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		_, err = prov.OpenStore("test_closing")
		require.NoError(t, err)

		delete(prov.dbs, "test_closing")

		err = prov.CloseStore("test_closing")
		require.NoError(t, err)
	})
}

func TestOpenStore(t *testing.T) {
	t.Run("use cached store", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		cstore := &couchDBStore{}
		prov.dbs["cached"] = cstore

		store, err := prov.OpenStore("cached")
		require.NoError(t, err)
		require.Equal(t, cstore, store)
	})
}

func TestOpenStoreFailures(t *testing.T) {
	t.Run("no name error", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		_, err = prov.OpenStore("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "")
	})
}

func TestCloseStore(t *testing.T) {
	t.Run("no config error", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		_, err = prov.OpenStore("test_closing")
		require.NoError(t, err)

		err = prov.CloseStore("test_closing")
		require.NoError(t, err)
	})

	t.Run("no mapping found", func(t *testing.T) {
		prov, err := NewProvider(couchDBURL)
		require.NoError(t, err)

		_, err = prov.OpenStore("test_closing")
		require.NoError(t, err)

		delete(prov.dbs, "test_closing")

		err = prov.CloseStore("test_closing")
		require.NoError(t, err)
	})
}
