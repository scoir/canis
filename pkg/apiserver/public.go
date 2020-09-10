package apiserver

import (
	"context"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
	"github.com/scoir/canis/pkg/indy/wrapper/crypto"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

func (r *APIServer) createAgentPublicDID(a *datastore.Agent) error {
	//TODO:  where is the methodName stored
	//TODO: use Indy IndyVDR for now but do NOT tie ourselves to Indy!
	did, err := r.didStore.GetPublicDID()
	if err != nil {
		return errors.Wrap(err, "unable to get public DID.")
	}

	endpoint, err := r.loadbalancer.GetEndpoint(context.Background(), &api.EndpointRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to retrieve endpoint for new agent")
	}

	mysig := crypto.NewSigner(did.KeyPair.RawPublicKey(), did.KeyPair.RawPrivateKey())

	agentPublicDID, agentPublicKeys, err := identifiers.CreateDID(&identifiers.MyDIDInfo{MethodName: "scr", Cid: true})
	if err != nil {
		return errors.Wrap(err, "unable to create agent DID")
	}

	err = r.client.CreateNym(agentPublicDID.DIDVal.MethodSpecificID, agentPublicDID.Verkey, vdr.EndorserRole, did.DID.DIDVal.MethodSpecificID, mysig)
	if err != nil {
		return errors.Wrap(err, "unable to set nym")
	}
	newDIDsig := crypto.NewSigner(agentPublicKeys.RawPublicKey(), agentPublicKeys.RawPrivateKey())

	err = r.client.SetEndpoint(agentPublicDID.DIDVal.MethodSpecificID, agentPublicDID.DIDVal.MethodSpecificID,
		endpoint.Endpoint, newDIDsig)
	if err != nil {
		return errors.Wrap(err, "unable to set endpoint")
	}

	a.PublicDID = &datastore.DID{
		DID: agentPublicDID,
		KeyPair: &datastore.KeyPair{
			PublicKey:  agentPublicKeys.PublicKey(),
			PrivateKey: agentPublicKeys.PrivateKey(),
		},
		Endpoint: endpoint.Endpoint,
	}

	return nil
}
