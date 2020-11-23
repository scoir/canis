/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package indy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
)

const (
	schemaV1 = "https://w3id.org/did/v1"
	keyType  = "Ed25519VerificationKey2018"
)

func (r *VDRI) Read(did string, opts ...vdriapi.ResolveOpts) (*diddoc.Doc, error) {
	if !strings.HasPrefix(did, fmt.Sprintf("did:%s", r.methodName)) {
		did = fmt.Sprintf("did:%s:%s", r.methodName, did)
	}
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

	if rply.Data == nil {
		return nil, errors.New("did not found")
	}

	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(rply.Data.(string)), &m)
	if err != nil {
		return nil, err
	}

	//TODO: support multiple pubkeys
	txnTime := time.Unix(int64(rply.TxnTime), 0)
	verkey, _ := m["verkey"].(string)
	pubKeyValue := base58.Decode(verkey)

	KID, err := localkms.CreateKID(pubKeyValue, kms.ED25519Type)

	pubKey := diddoc.NewPublicKeyFromBytes("#"+KID, keyType, "#id", pubKeyValue)
	verMethod := diddoc.NewReferencedVerificationMethod(pubKey, diddoc.Authentication, true)

	var svc []diddoc.Service
	serviceEndpoint, err := r.getEndpoint(parsedDID.MethodSpecificID)
	if err == nil {
		s := diddoc.Service{
			ID:              "#agent",
			Type:            vdriapi.DIDCommServiceType,
			ServiceEndpoint: serviceEndpoint,
			Priority:        0,
			RecipientKeys:   []string{verkey},
		}

		svc = append(svc, s)
	}

	doc := &diddoc.Doc{
		Context:        []string{schemaV1},
		ID:             did,
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
