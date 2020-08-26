/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/controller/webnotifier"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/datastore"
	ndid "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/runtime"
	"github.com/scoir/canis/pkg/schema"
)

//go:generate wire

type Steward struct {
	didcl     *didexchange.Client
	bouncer   ndid.Bouncer
	schemacl  *schema.Client
	notifier  *webnotifier.WebNotifier
	store     datastore.Store
	exec      runtime.Executor
	publicDID *datastore.DID
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	Datastore() (datastore.Store, error)
	Executor() (runtime.Executor, error)
	GetDIDClient() (*didexchange.Client, error)
	GetCredentialClient() (*issuecredential.Client, error)
	GetSchemaClient() (*schema.Client, error)
	GetSupervisor(h credential.Handler) (*credential.Supervisor, error)
	GetBouncer() (ndid.Bouncer, error)
}

func New(ctx provider) (*Steward, error) {

	var err error
	r := &Steward{}

	r.schemacl, _ = ctx.GetSchemaClient()
	store, err := ctx.Datastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access datastore")
	}

	r.store = store

	exec, err := ctx.Executor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access runtime executor")
	}

	r.exec = exec

	r.didcl, err = ctx.GetDIDClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create didexchange client in steward init")
	}

	_, err = ctx.GetSupervisor(r)
	if err != nil {
		return nil, err
	}

	r.bouncer, err = ctx.GetBouncer()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create did supervisor for high school agent")
	}

	err = r.bootstrap()
	if err != nil {
		return nil, errors.Wrap(err, "error bootstraping steward")
	}
	return r, nil
}

func (r *Steward) bootstrap() error {
	log.Println("Retrieving public did for Steward")

	did, err := r.store.GetPublicDID()
	if err != nil {
		return errors.New("Steward needs public did")
	}

	r.publicDID = did
	log.Printf("Public did is %s with verkey %s\n", r.publicDID.DID, r.publicDID.Verkey)
	return nil
}
