/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
)

var (
	psqlInfo = &Config{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "mysecretpassword",
		DB:       "test",
		SSLMode:  "disable",
	}
)

const (
	sqlStoreDBURL = "postgres://postgres:mysecretpassword@localhost:5432/?sslmode=disable"
)

// For these unit tests to run, you must ensure you have a Postgres DB instance running at the URL specified in
// sqlStoreDBURL.
// To run the tests manually, start an instance by running the following command in the terminal
// docker run -p 5432:5432 --name PostgresStoreTest -e POSTGRES_PASSWORD=mysecretpassword -d postgres:11.8
// delete using
//   docker kill PostgresStoreTest
//   docker rm PostgresStoreTest
func TestMain(m *testing.M) {
	err := waitForSQLDBToStart()
	if err != nil {
		fmt.Printf(err.Error() +
			". Make sure you start a sqlStoreDB instance using" +
			" 'docker run -p 5432:5432 postgres:11.8' before running the unit tests")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func waitForSQLDBToStart() error {
	db, err := sql.Open("postgres", sqlStoreDBURL)
	if err != nil {
		return err
	}

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: couldn't reach sql db server")
		default:
			err := db.Ping()
			if err != nil {
				return err
			}

			return nil
		}
	}
}

func TestSQLDBStore(t *testing.T) {
	t.Run("Test sql db store open close", func(t *testing.T) {
		prov, err := NewProvider(psqlInfo)
		require.NoError(t, err)

		pool, err := pgxpool.Connect(context.Background(), prov.adminURL)
		require.NoError(t, err)

		// query the postgres db to see if the store exists
		statement := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'test');`

		row := pool.QueryRow(context.Background(), statement)

		var exists bool
		err = row.Scan(&exists)
		require.NoError(t, err)

		if exists == false {
			statement = fmt.Sprintf("CREATE DATABASE test;")
			_, err = pool.Exec(context.Background(), statement)
			require.NoError(t, err)
		}

		_, err = prov.OpenStore("test_store")
		require.NoError(t, err)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestInsertListDID(t *testing.T) {
	t.Run("Test insert / list public did", func(t *testing.T) {
		prov, err := NewProvider(psqlInfo)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_list")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{
			DID:    "a did",
			Public: true,
		})
		require.NoError(t, err)

		didlist, err := store.ListDIDs(nil)
		require.NoError(t, err)

		require.Equal(t, didlist.Count, 1)

		did := didlist.DIDs[0]
		require.Equal(t, "a did", did.DID)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestSetGetPublicDID(t *testing.T) {
	t.Run("Test get / set public did", func(t *testing.T) {
		prov, err := NewProvider(psqlInfo)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_dids")
		require.NoError(t, err)

		err = store.InsertDID(&datastore.DID{DID: "did to be public", Public: false})
		require.NoError(t, err)
		err = store.InsertDID(&datastore.DID{DID: "another did", Public: true})
		require.NoError(t, err)

		err = store.SetPublicDID("did to be public")
		require.NoError(t, err)

		public, err := store.GetPublicDID()
		require.NoError(t, err)

		require.Equal(t, "did to be public", public.DID)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestSchema(t *testing.T) {
	t.Run("Test schema CRUD", func(t *testing.T) {
		prov, err := NewProvider(psqlInfo)
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

		list, err = store.ListSchema(nil)
		require.NoError(t, err)
		require.Equal(t, 1, list.Count)

		err = prov.Close()
		require.NoError(t, err)
	})
}

func TestAgent(t *testing.T) {
	t.Run("Test Agent CRUD", func(t *testing.T) {
		prov, err := NewProvider(psqlInfo)
		require.NoError(t, err)

		store, err := prov.OpenStore("test_agents")
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{ID: "agent id", Name: "an agent"})
		require.NoError(t, err)

		_, err = store.InsertAgent(&datastore.Agent{ID: "agent id 2", Name: "another agent", InvitationID: "abc123"})
		require.NoError(t, err)

		list, err := store.ListAgent(nil)
		require.NoError(t, err)
		require.Equal(t, 2, list.Count)

		err = store.UpdateAgent(&datastore.Agent{ID: "agent id", Name: "an different agent"})
		require.NoError(t, err)

		updated, err := store.GetAgent("agent id")
		require.NoError(t, err)
		require.Equal(t, "an different agent", updated.Name)

		invited, err := store.GetAgentByInvitation("abc123")
		require.NoError(t, err)
		require.Equal(t, "another agent", invited.Name)

		err = store.DeleteAgent("agent id")
		require.NoError(t, err)

		list, err = store.ListAgent(nil)
		require.NoError(t, err)
		require.Equal(t, 1, list.Count)

		err = prov.Close()
		require.NoError(t, err)
	})
}
