/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
)

const (
	datastoreKey = "datastore"
)

func (r *Provider) Datastore() (datastore.Provider, error) {
	dc := &framework.DatastoreConfig{}
	err := r.vp.UnmarshalKey(datastoreKey, dc)
	if err != nil {
		return nil, errors.Wrap(err, "execution environment is not correctly configured")
	}

	return r.datastoreMgr.StorageProvider(dc)
}
