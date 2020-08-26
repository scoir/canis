package apiserver

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/apiserver/mocks"
	"github.com/scoir/canis/pkg/datastore"
	dsstore "github.com/scoir/canis/pkg/datastore/mocks"
)

func TestNew(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		p := &mocks.Provider{}
		ds := &dsstore.Store{}

		p.On("GetSchemaClient").Return(nil, nil)
		p.On("Datastore").Return(ds, nil)
		p.On("Executor").Return(nil, nil)
		p.On("GetDIDClient").Return(nil, nil)
		p.On("GetSupervisor", mock.AnythingOfType("*steward.APIServer")).Return(nil, nil)
		p.On("GetBouncer").Return(nil, nil)
		ds.On("GetPublicDID").Return(&datastore.DID{}, nil)

		steward, err := New(p)
		assert.Nil(t, err)
		assert.NotNil(t, steward)
		p.AssertExpectations(t)
		ds.AssertExpectations(t)

	})
	t.Run("no DID", func(t *testing.T) {
		p := &mocks.Provider{}
		ds := &dsstore.Store{}
		p.On("GetSchemaClient").Return(nil, nil)
		p.On("Datastore").Return(ds, nil)
		p.On("Executor").Return(nil, nil)
		p.On("GetDIDClient").Return(nil, nil)
		p.On("GetSupervisor", mock.AnythingOfType("*steward.APIServer")).Return(nil, nil)
		p.On("GetBouncer").Return(nil, nil)
		ds.On("GetPublicDID").Return(nil, errors.New("boom"))

		steward, err := New(p)
		require.Error(t, err)
		require.Nil(t, steward)

		p.AssertExpectations(t)
		ds.AssertExpectations(t)
	})
}
