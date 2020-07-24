/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package mongodb

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	sqlStoreDBURL = "postgres://postgres:mysecretpassword@localhost:5432/?sslmode=disable"
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