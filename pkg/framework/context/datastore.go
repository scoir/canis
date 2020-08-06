/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mongodb"
	"github.com/scoir/canis/pkg/datastore/postgres"
)

const (
	datastoreKey = "datastore"
)

type DatastoreConfig struct {
	Database string               `mapstructure:"database"`
	Mongo    *mongodb.Config      `mapstructure:"mongo"`
	Postgres *postgres.Config     `mapstructure:"postgres"`
	CouchDB  *couchdbstore.Config `mapstructure:"couchdb"`
}

func (r *Provider) Datastore() (datastore.Provider, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.ds != nil {
		return r.ds, nil
	}

	dc, err := r.GetDatastoreConfig()
	if err != nil {
		return nil, err
	}

	switch dc.Database {
	case "mongo":
		r.ds, err = mongodb.NewProvider(dc.Mongo)
	case "postgres":
		r.ds, err = postgres.NewProvider(dc.Postgres)
	case "couchdb":
		r.ds, err = couchdbstore.NewProvider(dc.CouchDB.URL)
	default:
		return nil, errors.New("no datastore configuration was provided")
	}

	return r.ds, errors.Wrap(err, "unable to get datastore from config")
}
