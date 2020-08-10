/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"fmt"

	couchdbstore "github.com/scoir/canis/pkg/datastore/couchdb"
	"github.com/scoir/canis/pkg/datastore/mongodb"
	"github.com/scoir/canis/pkg/datastore/postgres"
)

type Endpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (r Endpoint) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type DatastoreConfig struct {
	Database string               `mapstructure:"database"`
	Mongo    *mongodb.Config      `mapstructure:"mongo"`
	Postgres *postgres.Config     `mapstructure:"postgres"`
	CouchDB  *couchdbstore.Config `mapstructure:"couchdb"`
}

func (r *DatastoreConfig) Name() string {
	switch r.Database {
	case "mongo":
		return r.Mongo.Database
	case "postgres":
		return r.Postgres.DB
	case "couchdb":
		return r.CouchDB.URL
	default:
		return ""
	}
}

func (r *DatastoreConfig) SetName(n string) {
	switch r.Database {
	case "mongo":
		r.Mongo.Database = n
	case "postgres":
		r.Postgres.DB = n
	case "couchdb":
		r.CouchDB.URL = n
	}
}
