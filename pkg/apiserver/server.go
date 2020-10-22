/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	cengine "github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
	doorman "github.com/scoir/canis/pkg/didcomm/doorman/api/protogen"
	api "github.com/scoir/canis/pkg/didcomm/issuer/api/protogen"
	lbapi "github.com/scoir/canis/pkg/didcomm/loadbalancer/api/protogen"
	verifier "github.com/scoir/canis/pkg/didcomm/verifier/api/protogen"
	"github.com/scoir/canis/pkg/indy"
	pengine "github.com/scoir/canis/pkg/presentproof/engine"
)

type APIServer struct {
	keyMgr               kms.KeyManager
	agentStore           datastore.Store
	schemaStore          datastore.Store
	store                datastore.Store
	client               vdrClient
	schemaRegistry       cengine.CredentialRegistry
	presentationRegistry pengine.PresentationRegistry
	doorman              doorman.DoormanClient
	issuer               api.IssuerClient
	verifier             verifier.VerifierClient
	loadbalancer         lbapi.LoadbalancerClient
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	KMS() kms.KeyManager
	Store() datastore.Store
	IndyVDR() (indy.IndyVDRClient, error)
	GetDoormanClient() (doorman.DoormanClient, error)
	GetIssuerClient() (api.IssuerClient, error)
	GetVerifierClient() (verifier.VerifierClient, error)
	GetLoadbalancerClient() (lbapi.LoadbalancerClient, error)
	GetCredentialEngineRegistry() (cengine.CredentialRegistry, error)
	GetPresentationEngineRegistry() (pengine.PresentationRegistry, error)
}

type vdrClient interface {
	SetEndpoint(did, from string, ep string, signer vdr.Signer) error
	CreateNym(did, verkey, role, from string, signer vdr.Signer) error
	GetNym(did string) (*vdr.ReadReply, error)
}

func New(ctx provider) (*APIServer, error) {

	var err error
	r := &APIServer{}

	r.keyMgr = ctx.KMS()
	r.schemaStore = ctx.Store()
	r.agentStore = ctx.Store()
	r.store = ctx.Store()

	r.client, err = ctx.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get IndyVDR")
	}

	r.doorman, err = ctx.GetDoormanClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get doorman client")
	}

	r.issuer, err = ctx.GetIssuerClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get issuer client")
	}

	r.verifier, err = ctx.GetVerifierClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get verifier client")
	}

	r.loadbalancer, err = ctx.GetLoadbalancerClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get loadbalancer client")
	}

	r.schemaRegistry, err = ctx.GetCredentialEngineRegistry()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load credential engine registry")
	}

	r.presentationRegistry, err = ctx.GetPresentationEngineRegistry()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load presentation engine registry")
	}

	return r, nil
}
