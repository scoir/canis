/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/runtime"
)

//go:generate wire
type APIServer struct {
	agentStore  datastore.Store
	schemaStore datastore.Store
	didStore    datastore.Store
	exec        runtime.Executor
	client      *vdr.Client
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	StorageProvider() (datastore.Provider, error)
	Executor() (runtime.Executor, error)
}

func New(ctx provider) (*APIServer, error) {

	var err error
	r := &APIServer{}

	storageProvider, err := ctx.StorageProvider()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access datastore")
	}

	r.schemaStore, err = storageProvider.OpenStore("Schema")
	if err != nil {
		return nil, err
	}

	r.agentStore, err = storageProvider.OpenStore("Agent")
	if err != nil {
		return nil, err
	}

	r.didStore, err = storageProvider.OpenStore("DID")
	if err != nil {
		return nil, err
	}

	exec, err := ctx.Executor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access runtime executor")
	}

	r.exec = exec

	return r, nil
}
