/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package framework

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLedgerStoreConfig(t *testing.T) {
	t.Run("test address works", func(t *testing.T) {
		lsc := &LedgerStoreConfig{
			Database: "",
		}

		dp, err := lsc.StorageProvider()
		require.Error(t, err)
		require.Contains(t, err.Error(), "no ledgerstore configuration was provided")
		require.Nil(t, dp)
	})
}
