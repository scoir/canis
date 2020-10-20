package apiserver

import (
	"context"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
)

func (r *APIServer) createAgentPublicDID(a *datastore.Agent) error {
	//TODO:  where is the methodName stored
	//TODO: use Indy IndyVDR for now but do NOT tie ourselves to Indy!
	did, err := r.store.GetPublicDID()
	if err != nil {
		return errors.Wrap(err, "unable to get public DID.")
	}

	mysig, err := r.getSignerForID(did.KeyPair.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get signer for public DID")
	}

	endpoint, err := r.loadbalancer.GetEndpoint(context.Background(), &api.EndpointRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to retrieve endpoint for new agent")
	}

	newKeyID, pubKey, err := r.keyMgr.CreateAndExportPubKeyBytes(kms.ED25519Type)
	if err != nil {
		return errors.Wrap(err, "unable to create and export public key")
	}

	newDIDsig, err := r.getSignerForID(newKeyID)
	if err != nil {
		return errors.Wrap(err, "unable to load signer primitives")
	}

	agentPublicDID, err := identifiers.CreateDID(&identifiers.MyDIDInfo{PublicKey: pubKey, MethodName: "scr", Cid: true})
	if err != nil {
		return errors.Wrap(err, "unable to create agent DID")
	}

	err = r.client.CreateNym(agentPublicDID.DIDVal.MethodSpecificID, agentPublicDID.Verkey, vdr.EndorserRole, did.DID.DIDVal.MethodSpecificID, mysig)
	if err != nil {
		return errors.Wrap(err, "unable to set nym")
	}

	err = r.client.SetEndpoint(agentPublicDID.DIDVal.MethodSpecificID, agentPublicDID.DIDVal.MethodSpecificID,
		endpoint.Endpoint, newDIDsig)
	if err != nil {
		return errors.Wrap(err, "unable to set endpoint")
	}

	a.PublicDID = &datastore.DID{
		DID: agentPublicDID,
		KeyPair: &datastore.KeyPair{
			ID:        newKeyID,
			PublicKey: base58.Encode(pubKey),
		},
		Endpoint: endpoint.Endpoint,
	}

	return nil
}

func (r *APIServer) getSignerForID(kid string) (vdr.Signer, error) {
	kh, err := r.keyMgr.Get(kid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get private key")
	}

	privKeyHandle := kh.(*keyset.Handle)
	prim, err := privKeyHandle.Primitives()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load signer primitives")
	}
	sig := prim.Primary.Primitive.(*subtle.ED25519Signer)
	return sig, nil
}
