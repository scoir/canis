package engine

import (
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
)

type providerMock struct {
	store *mocks.Store
}

func (r *providerMock) Store() datastore.Store {
	return r.store
}

func NewProvider() *providerMock {
	return &providerMock{store: &mocks.Store{}}
}
