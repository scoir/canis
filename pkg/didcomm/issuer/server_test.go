/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"testing"

	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
)

func _TestProviderFailures(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, err := New(&MockProvider{})
		require.NoError(t, err)
		require.NotNil(t, s)
	})
}

type MockProvider struct {
	didErr   error
	issueErr error
}

func (r *MockProvider) GetAriesContext() (*ariescontext.Provider, error) {
	return nil, nil
}

func (r *MockProvider) GetDatastore() datastore.Store {
	return &mocks.Store{}
}
