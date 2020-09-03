/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	doorman "github.com/scoir/canis/pkg/didcomm/doorman/api"
	issuer "github.com/scoir/canis/pkg/didcomm/issuer/api"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

type APIServer struct {
	agentStore   datastore.Store
	schemaStore  datastore.Store
	didStore     datastore.Store
	storeManager *manager.DataProviderManager
	client       *vdr.Client

	doorman     doorman.DoormanClient
	issuer      issuer.IssuerClient
	watcherLock sync.RWMutex
	watchers    []chan *api.AgentEvent
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	StorageManager() *manager.DataProviderManager
	Store() datastore.Store
	VDR() (*vdr.Client, error)
	GetDoormanClient() (doorman.DoormanClient, error)
	GetIssuerClient() (issuer.IssuerClient, error)
}

func New(ctx provider) (*APIServer, error) {

	var err error
	r := &APIServer{
		watchers: make([]chan *api.AgentEvent, 0),
	}

	r.schemaStore = ctx.Store()
	r.agentStore = ctx.Store()
	r.didStore = ctx.Store()

	r.client, err = ctx.VDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get VDR")
	}

	r.doorman, err = ctx.GetDoormanClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get doorman client")
	}

	r.issuer, err = ctx.GetIssuerClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get issuer client")
	}

	r.storeManager = ctx.StorageManager()

	return r, nil
}
