package mock

import (
	"github.com/hyperledger/aries-framework-go/pkg/storage"
)

//go:generate mockery -inpkg -name=Provider
type Provider interface {
	// OpenStore opens a store with given name space and returns the handle
	OpenStore(name string) (storage.Store, error)

	// CloseStore closes store of given name space
	CloseStore(name string) error

	// Close closes all stores created under this store provider
	Close() error
}

//go:generate mockery -inpkg -name=Store
type Store interface {
	// Put stores the key and the record
	Put(k string, v []byte) error

	// Get fetches the record based on key
	Get(k string) ([]byte, error)

	// Iterator returns an iterator for the latest snapshot of the
	// underlying store
	//
	// Args:
	//
	// startKey: Start of the key range, include in the range.
	// endKey: End of the key range, not include in the range.
	//
	// Returns:
	//
	// StoreIterator: iterator for result range
	Iterator(startKey, endKey string) storage.StoreIterator

	// Delete will delete a record with k key
	Delete(k string) error
}
