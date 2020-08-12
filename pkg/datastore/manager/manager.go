package manager

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	couchdbstore "github.com/scoir/canis/pkg/datastore/couchdb"
	"github.com/scoir/canis/pkg/datastore/mongodb"
	"github.com/scoir/canis/pkg/datastore/postgres"
	"github.com/scoir/canis/pkg/framework"
)

type DataProviderManager struct {
	lock sync.Mutex
	dc   *framework.DatastoreConfig
	ds   map[string]datastore.Provider
}

func NewDataProviderManager(dc *framework.DatastoreConfig) *DataProviderManager {
	return &DataProviderManager{
		dc: dc,
		ds: map[string]datastore.Provider{},
	}
}

func (r *DataProviderManager) Config() *framework.DatastoreConfig {
	return r.dc
}

func (r *DataProviderManager) DefaultStoreProvider() (datastore.Provider, error) {
	return r.StorageProvider(r.dc)
}

func (r *DataProviderManager) StorageProvider(dc *framework.DatastoreConfig) (datastore.Provider, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	key := fmt.Sprintf("%s:%s", dc.Database, dc.Name())
	ds, ok := r.ds[key]
	if ok {
		return ds, nil
	}

	var err error
	switch dc.Database {
	case "mongo":
		ds, err = mongodb.NewProvider(dc.Mongo)
	case "postgres":
		ds, err = postgres.NewProvider(dc.Postgres)
	case "couchdb":
		ds, err = couchdbstore.NewProvider(dc.CouchDB.URL)
	default:
		return nil, errors.New("no datastore configuration was provided")
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to create datastore based on config")
	}

	if r.ds != nil {
		r.ds[key] = ds
	}

	return ds, nil

}
