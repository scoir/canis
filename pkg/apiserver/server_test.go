package apiserver

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/scoir/canis/pkg/apiserver/mocks"
	dsstore "github.com/scoir/canis/pkg/datastore/mocks"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

func TestNew(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		p := &mocks.Provider{}
		dps := &dsstore.Provider{}
		s := &dsstore.Store{}

		p.On("VDR").Return(&vdr.Client{}, nil)
		p.On("StorageProvider").Return(dps, nil)
		p.On("StorageManager").Return(nil, nil)
		dps.On("OpenStore", "Schema").Return(s, nil)
		dps.On("OpenStore", "Agent").Return(s, nil)
		dps.On("OpenStore", "DID").Return(s, nil)

		steward, err := New(p)
		assert.Nil(t, err)
		assert.NotNil(t, steward)
		p.AssertExpectations(t)
		dps.AssertExpectations(t)

	})
}
