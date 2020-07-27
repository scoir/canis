package couchdbstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kivik/kivik"
)

const (
	couchDBURL = "localhost:5984"
)

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
