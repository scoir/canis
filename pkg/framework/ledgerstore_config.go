/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	couchdbstore "github.com/hyperledger/aries-framework-go/pkg/storage/couchdb"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mysql"
	"github.com/pkg/errors"
)

type LedgerStoreConfig struct {
	Database string
	URL      string
}

func (r *LedgerStoreConfig) StorageProvider() (storage.Provider, error) {
	var sp storage.Provider
	var err error

	switch r.Database {
	case "mysql":
		sp, err = mysql.NewProvider(r.URL)
	case "couchdb":
		sp, err = couchdbstore.NewProvider(r.URL)
	default:
		return nil, errors.New("no ledgerstore configuration was provided")
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to create ledgerstore based on config")
	}

	return sp, nil
}
