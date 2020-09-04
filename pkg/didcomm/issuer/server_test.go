/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"testing"

	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/stretchr/testify/require"

	mockStore "github.com/scoir/canis/pkg/mock/storage"
)

func TestProviderFailures(t *testing.T) {
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
	panic("implement me")
}

func (r *MockProvider) GetStorageProvider() (storage.Provider, error) {
	return mockStore.NewMockStoreProvider(), nil
}