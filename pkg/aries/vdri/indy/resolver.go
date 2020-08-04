/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package indy

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
)

const (
	schemaV1    = "https://w3id.org/did/v1"
	keyType     = "Ed25519VerificationKey2018"
	serviceType = "IndyAgent"
)

func (r *VDRI) Read(did string, opts ...vdriapi.ResolveOpts) (*diddoc.Doc, error) {
	parsedDID, err := diddoc.Parse(did)
	if err != nil {
		return nil, fmt.Errorf("parsing did failed in indy resolver: (%w)", err)
	}

	if parsedDID.Method != r.methodName {
		return nil, fmt.Errorf("invalid indy method ID: %s", parsedDID.MethodSpecificID)
	}

	resOpts := &vdriapi.ResolveDIDOpts{}
	for _, opt := range opts {
		opt(resOpts)
	}

	rply, err := r.client.GetNym(parsedDID.MethodSpecificID)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(rply.Data.(string)), &m)
	if err != nil {
		return nil, err
	}

	//TODO: support multiple pubkeys
	txnTime := time.Unix(int64(rply.TxnTime), 0)
	didKey, _ := m["dest"].(string)
	verkey, _ := m["verkey"].(string)
	pubKeyValue := base58.Decode(verkey)
	keyID := fmt.Sprintf("%s#0", didKey)
	pubKey := diddoc.NewPublicKeyFromBytes(keyID, keyType, didKey, pubKeyValue)
	verMethod := diddoc.NewReferencedVerificationMethod(pubKey, diddoc.Authentication, true)

	var svc []diddoc.Service
	serviceEndpoint, err := r.getEndpoint(parsedDID.MethodSpecificID)
	if err == nil {
		s := diddoc.Service{
			ID:              "#agent",
			Type:            serviceType,
			ServiceEndpoint: serviceEndpoint,
			Priority:        0,
			RecipientKeys:   []string{keyID},
		}

		svc = append(svc, s)
	}

	doc := &diddoc.Doc{
		Context:        []string{schemaV1},
		ID:             didKey,
		PublicKey:      []diddoc.PublicKey{*pubKey},
		Authentication: []diddoc.VerificationMethod{*verMethod},
		Service:        svc,
		Created:        &txnTime,
		Updated:        &txnTime,
	}

	return doc, nil
}

func (r *VDRI) getEndpoint(did string) (string, error) {
	rply, err := r.client.GetEndpoint(did)
	if err != nil || rply.Data == nil {
		return "", errors.New("not found")
	}

	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(rply.Data.(string)), &m)
	if err != nil {
		return "", err
	}

	mm, ok := m["endpoint"].(map[string]interface{})
	if !ok {
		return "", errors.New("not found")
	}

	ep, ok := mm["endpoint"].(string)
	if !ok {
		return "", errors.New("not found")
	}

	return ep, nil
}
