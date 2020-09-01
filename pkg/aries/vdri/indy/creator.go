/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package indy

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
)

func (r *VDRI) Build(pubKey *vdriapi.PubKey, opts ...vdriapi.DocOpts) (*diddoc.Doc, error) {

	if pubKey.Type != keyType {
		return nil, fmt.Errorf("only %s key type supported", keyType)
	}

	docOpts := &vdriapi.CreateDIDOpts{}
	for _, opt := range opts {
		opt(docOpts)
	}

	pubKeyValue := base58.Decode(string(pubKey.Value))
	methodID := base58.Encode(pubKeyValue[0:16])
	didKey := fmt.Sprintf("did:%s:%s", r.methodName, methodID)
	keyID := fmt.Sprintf("%s#%s", didKey, methodID)
	publicKey := did.NewPublicKeyFromBytes(keyID, keyType, didKey, pubKeyValue)

	var service []diddoc.Service
	if docOpts.ServiceType != "" {
		s := diddoc.Service{
			ID:              "#agent",
			Type:            docOpts.ServiceType,
			ServiceEndpoint: docOpts.ServiceEndpoint,
		}

		if docOpts.ServiceType == vdriapi.DIDCommServiceType {
			s.RecipientKeys = []string{string(pubKey.Value)}
			s.Priority = 0
		}

		service = append(service, s)
	}

	// Created/Updated time
	t := time.Now()
	doc := diddoc.BuildDoc(
		diddoc.WithService(service),
		diddoc.WithCreatedTime(t),
		diddoc.WithUpdatedTime(t),
	)

	doc.ID = didKey
	// Create a did doc based on the mandatory value: publicKeys & authentication
	doc.PublicKey = []diddoc.PublicKey{*publicKey}
	doc.Authentication = []did.VerificationMethod{
		{PublicKey: *publicKey},
	}

	return doc, nil
}
