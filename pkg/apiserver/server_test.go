package apiserver

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/apiserver/mocks"
	dsstore "github.com/scoir/canis/pkg/datastore/mocks"
)

func TestNew(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		p := &mocks.Provider{}
		ds := &dsstore.Store{}

		p.On("Store").Return(ds, nil).Times(3)
		p.On("IndyVDR").Return(nil, nil)
		p.On("GetDoormanClient").Return(nil, nil)
		p.On("GetMediatorClient").Return(nil, nil)
		p.On("GetIssuerClient").Return(nil, nil)
		p.On("GetVerifierClient").Return(nil, nil)
		p.On("GetLoadbalancerClient").Return(nil, nil)
		p.On("GetCredentialEngineRegistry").Return(nil, nil)
		p.On("GetPresentationEngineRegistry").Return(nil, nil)
		p.On("KMS").Return(nil, nil)
		p.On("MediatorKMS").Return(nil, nil)

		server, err := New(p)
		require.Nil(t, err)
		require.NotNil(t, server)
		p.AssertExpectations(t)
		ds.AssertExpectations(t)

	})
	t.Run("with error", func(t *testing.T) {
		p := &mocks.Provider{}
		ds := &dsstore.Store{}

		p.On("Store").Return(ds, nil).Times(3)
		p.On("KMS").Return(nil, nil)
		p.On("MediatorKMS").Return(nil, nil)
		p.On("IndyVDR").Return(nil, errors.New("Boom"))

		steward, err := New(p)
		require.Error(t, err)
		require.Nil(t, steward)

		p.AssertExpectations(t)
		ds.AssertExpectations(t)
	})
}
