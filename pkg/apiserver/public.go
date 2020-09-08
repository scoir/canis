package apiserver

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
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

	endpoint := "http://420.69.420.69:6969"

	mysig := crypto.NewSigner(did.KeyPair.RawPublicKey(), did.KeyPair.RawPrivateKey())
	fmt.Println("Steward DID:", did.DID)
	fmt.Println("Steward Verkey:", did.DID.AbbreviateVerkey())

	agentPublicDID, agentPublicKeys, err := identifiers.CreateDID(&identifiers.MyDIDInfo{MethodName: "scr", Cid: true})
	if err != nil {
		return errors.Wrap(err, "unable to create agent DID")
	}

	err = r.client.CreateNym(agentPublicDID.DIDVal.MethodSpecificID, agentPublicDID.Verkey, vdr.EndorserRole, did.DID.DIDVal.MethodSpecificID, mysig)
	if err != nil {
		return errors.Wrap(err, "unable to set nym")
	}
	fmt.Println("New Endorser DID:", agentPublicDID.String())
	fmt.Println("New Endorser Verkey:", agentPublicDID.AbbreviateVerkey())
	fmt.Println("Place These in Wallet:")
	fmt.Println("Public:", agentPublicKeys.PublicKey())
	fmt.Println("Private:", agentPublicKeys.PrivateKey())

	newDIDsig := crypto.NewSigner(agentPublicKeys.RawPublicKey(), agentPublicKeys.RawPrivateKey())

	err = r.client.SetEndpoint(agentPublicDID.DIDVal.MethodSpecificID, agentPublicDID.DIDVal.MethodSpecificID, endpoint, newDIDsig)
	if err != nil {
		return errors.Wrap(err, "unable to set endpoint")
	}

	a.PublicDID = &datastore.DID{
		DID: agentPublicDID,
		KeyPair: &datastore.KeyPair{
			PublicKey:  agentPublicKeys.PublicKey(),
			PrivateKey: agentPublicKeys.PrivateKey(),
		},
		Endpoint: endpoint,
	}

	return nil
}
