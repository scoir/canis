/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mongodb"
)

type DatastoreConfig struct {
	Database string          `mapstructure:"database"`
	Mongo    *mongodb.Config `mapstructure:"mongo"`
}

func (r *DatastoreConfig) StorageProvider() (datastore.Provider, error) {
	var dp datastore.Provider
	var err error

	switch r.Database {
	case "mongo":
		dp, err = mongodb.NewProvider(r.Mongo)

	default:
		return nil, errors.New("no datastore configuration was provided")
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to create datastore based on config")
	}

	return dp, nil
}
