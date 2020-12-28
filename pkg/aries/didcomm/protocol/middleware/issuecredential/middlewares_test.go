/*
Copyright Scoir, Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuecredential

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
)

func TestSaveCredentials(t *testing.T) {
	t.Run("request-receive", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		msg := service.DIDCommMsgMap{
			"@id": "123",
			"~thread": map[string]interface{}{
				"thid": "abc",
			},
		}
		cred := &datastore.IssuedCredential{
			ID:          "cred-1",
			SystemState: "request-received",
		}

		h := &handler{}

		md := &MockMetadata{}
		md.On("StateName").Return("request-receive")
		md.On("Message").Return(msg)
		suite.store.On("FindCredentialByProtocolID", "abc").Return(cred, nil)
		suite.store.On("UpdateCredential", cred).Return(nil)

		err := suite.target(h).Handle(md)
		require.NoError(t, err)
		require.Equal(t, md, h.md)
	})
	t.Run("unable to update credential", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		msg := service.DIDCommMsgMap{
			"@id": "123",
			"~thread": map[string]interface{}{
				"thid": "abc",
			},
		}
		cred := &datastore.IssuedCredential{
			ID:          "cred-1",
			SystemState: "request-received",
		}

		h := &handler{}

		md := &MockMetadata{}
		md.On("StateName").Return("request-receive")
		md.On("Message").Return(msg)
		suite.store.On("FindCredentialByProtocolID", "abc").Return(cred, nil)
		suite.store.On("UpdateCredential", cred).Return(errors.New("unable to update"))

		err := suite.target(h).Handle(md)
		require.Error(t, err)
		require.Equal(t, "unexpected error updating issued credential cred-1: unable to update", err.Error())
	})
	t.Run("credential not found", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		msg := service.DIDCommMsgMap{
			"@id": "123",
			"~thread": map[string]interface{}{
				"thid": "abc",
			},
		}

		h := &handler{}

		md := &MockMetadata{}
		md.On("StateName").Return("request-receive")
		md.On("Message").Return(msg)
		suite.store.On("FindCredentialByProtocolID", "abc").Return(nil, errors.New("not found"))

		err := suite.target(h).Handle(md)
		require.Error(t, err)
		require.Equal(t, "unable to find cred with ID abc: (not found)", err.Error())
	})
	t.Run("offer-sent", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		h := &handler{}
		md := &MockMetadata{}
		md.On("StateName").Return("offer-sent")
		err := suite.target(h).Handle(md)
		require.NoError(t, err)
	})

}

type suite struct {
	target issuecredential.Middleware
	store  *mocks.Store
}

type handler struct {
	Err error
	md  issuecredential.Metadata
}

func (r *handler) Handle(metadata issuecredential.Metadata) error {
	r.md = metadata
	return r.Err
}

func setup(t *testing.T) (*suite, func()) {
	s := &suite{
		store: &mocks.Store{},
	}

	p := &MockProvider{}
	p.On("Store").Return(s.store)

	s.target = SaveCredentials(p)
	p.AssertExpectations(t)

	return s, func() {
		s.store.AssertExpectations(t)
	}
}
