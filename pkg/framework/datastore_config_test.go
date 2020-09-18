/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package framework

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDataStoreConfig(t *testing.T) {
	t.Run("test address works", func(t *testing.T) {
		dsc := &DatastoreConfig{
			Database: "",
		}

		dp, err := dsc.StorageProvider()
		require.Error(t, err)
		require.Contains(t, err.Error(), "no datastore configuration was provided")
		require.Nil(t, dp)
	})
}
