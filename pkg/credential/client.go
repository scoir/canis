/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
)

type Client struct {
	store storage.Store
}

type credKeys struct {
	PrivKey             json.RawMessage
	KeyCorrectnessProof json.RawMessage
	SchemaID            string
}

func NewClient() (*Client, error) {
	//provider := conf.GetAriesContext()
	//store, err := provider.StorageProvider().OpenStore("cred-defs")
	//if err != nil {
	//	return nil, errors.Wrap(err, "can't get storage for credential client")
	//}
	return &Client{
		store: nil,
	}, nil
}

//func (r *Client) CreateCredentialDefinition(issuer, verkey string, schema *indy.SchemaData) (string, error) {
//
//	indycd := anoncreds.NewCredentailDefinition()
//	indycd.AddSchemaFields(schema.AttrNames...)
//	err := indycd.Finalize()
//	if err != nil {
//		return "", errors.Wrap(err, "unable to finalize indy credential definition")
//	}
//	pubd, _ := indycd.PublicKey()
//	pubKeyDef := map[string]interface{}{}
//
//	err = json.Unmarshal([]byte(pubd), &pubKeyDef)
//	if err != nil {
//		return "", errors.Wrap(err, "invalid cl pubkey")
//	}
//
//	pubKey, _ := pubKeyDef["p_key"].(map[string]interface{})
//	credDefId, err := r.vdr.CreateClaimDef(issuer, verkey, schema.SeqNo, pubKey, nil)
//	if err != nil {
//		return "", errors.Wrap(err, "unable to create claim def")
//	}
//
//	pk, _ := indycd.PrivateKey()
//	kcp, _ := indycd.KeyCorrectnessProof()
//	ck := credKeys{
//		SchemaID:            schema.ID,
//		PrivKey:             []byte(pk),
//		KeyCorrectnessProof: []byte(kcp),
//	}
//
//	d, _ := json.Marshal(ck)
//	log.Printf("putting %s cred def in wallet\n", credDefId)
//	err = r.store.Put(credDefId, d)
//	if err != nil {
//		return "", errors.Wrap(err, "unable to save cred def")
//	}
//
//	return credDefId, nil
//}
//
//func (r *Client) GetCredentialDefinition(did string, s *indy.SchemaData) ([]string, error) {
//	credDef, err := r.vdr.GetClaimDef(did, s.SeqNo)
//	if err != nil {
//		return nil, errors.Wrap(err, "unable to get claim def from vdr")
//	}
//
//	credDefResp := schema.CredentialDefinitionIDsResult{
//		CredentialDefinitionIDs: []string{credDef.ID},
//	}
//
//	return credDefResp.CredentialDefinitionIDs, nil
//}
//
//func (r *Client) CreateCredentialOffer(credDefID string) (*indy.CredentialOffer, error) {
//	out := &indy.CredentialOffer{
//		CredDefID: credDefID,
//	}
//
//	ck := credKeys{}
//	d, err := r.store.Get(credDefID)
//	if err != nil {
//		return nil, errors.Wrapf(err, "%s not found in wallet", credDefID)
//	}
//
//	err = json.Unmarshal(d, &ck)
//	if err != nil {
//		return nil, errors.Wrap(err, "invalid cred def type stored in wallet")
//	}
//
//	out.KeyCorrectnessProof = string(ck.KeyCorrectnessProof)
//	out.SchemaID = ck.SchemaID
//	out.Nonce, err = anoncreds.NewNonce()
//	if err != nil {
//		return nil, errors.Wrap(err, "error creating nonce")
//	}
//
//	return out, nil
//}
