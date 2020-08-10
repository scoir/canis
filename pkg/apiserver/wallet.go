package apiserver

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/crypto"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

func (r *APIServer) createAgentWallet(a *datastore.Agent) error {
	//TODO:  where is the methodName stored
	//TODO: use Indy VDR for now but do NOT tie ourselves to Indy!
	did, keyPair, err := identifiers.CreateDID(&identifiers.MyDIDInfo{Cid: true, MethodName: "sov"})
	if err != nil {
		return errors.Wrap(err, "unable to create DID")
	}

	mysig := crypto.NewSigner(keyPair.RawPublicKey(), keyPair.RawPrivateKey())

	fmt.Println("Steward DID:", did.String())
	fmt.Println("Steward Verkey:", did.Verkey)
	fmt.Println("Steward Short Verkey:", did.AbbreviateVerkey())

	someRandomDID, someRandomKP, err := identifiers.CreateDID(&identifiers.MyDIDInfo{MethodName: "sov", Cid: true})
	if err != nil {
		return errors.Wrap(err, "unable to create DID")
	}

	err = r.client.CreateNym(someRandomDID.DIDVal.DID, someRandomDID.Verkey, vdr.EndorserRole, did.DIDVal.DID, mysig)
	if err != nil {
		return errors.Wrap(err, "unable to create DID")
	}
	fmt.Println("New Endorser DID:", someRandomDID.String())
	fmt.Println("New Endorser Verkey:", someRandomDID.AbbreviateVerkey())
	fmt.Println("Place These in Wallet:")
	fmt.Println("Public:", someRandomKP.PublicKey())
	fmt.Println("Private:", someRandomKP.PrivateKey())

	newDIDsig := crypto.NewSigner(someRandomKP.RawPublicKey(), someRandomKP.RawPrivateKey())

	err = r.client.SetEndpoint(someRandomDID.DIDVal.DID, someRandomDID.DIDVal.DID, "http://420.69.420.69:6969", newDIDsig)
	if err != nil {
		return errors.Wrap(err, "unable to create DID")
	}

	return nil
}
