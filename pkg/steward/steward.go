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
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/datastore"
	ndid "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/runtime"
	"github.com/scoir/canis/pkg/schema"
)

//go:generate wire

type Steward struct {
	ctx       api.Provider
	didcl     *didexchange.Client
	bouncer   *ndid.Bouncer
	schemacl  *schema.Client
	credcl    *issuecredential.Client
	notifier  *webnotifier.WebNotifier
	store     datastore.Store
	exec      runtime.Executor
	publicDID *datastore.DID
}

func New(ctx api.Provider, conf *Config, sc *schema.Client) (*Steward, error) {

	var err error
	r := &Steward{
		ctx:      ctx,
		schemacl: sc,
	}

	store, err := conf.Datastore.Datastore()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access datastore")
	}

	r.store = store

	exec, err := conf.Execution.Executor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access runtime executor")
	}

	r.exec = exec

	r.didcl, err = conf.GetDIDClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create didexchange client in steward init")
	}

	r.credcl, err = conf.GetCredentialClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create issue credential client in steward init")
	}

	sup, err := credential.New(conf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create credential supervisor for steward")
	}
	err = sup.Start(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to start credential supervisor for steward")
	}

	r.bouncer, err = ndid.NewBouncer(conf)
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
		//TODO: This should be done with external tool!
		//log.Println("No public PeerDID, creating one")
		//
		//did, verkey, err := r.vdr.CreateNym()
		//if err != nil {
		//	return errors.Wrap(err, "creating public nym, steward/bootstrap")
		//}
		//
		//log.Printf("Going to use %s as did and %s as verkey\n", did, verkey)
		//err = r.ledgerBrowser.RegisterPublicDID(did, verkey, ScoirStewardAlias, ledger.StewardRole)
		//if err != nil {
		//	return errors.Wrap(err, "error registering public PeerDID in bootstrap")
		//}
		//
		//log.Println("PeerDID registered on Ledger as Steward and set as public with agent")
	} else {
		r.publicDID = did
		log.Printf("Public did is %s with verkey %s\n", r.publicDID.DID, r.publicDID.Verkey)
	}

	return nil
}
