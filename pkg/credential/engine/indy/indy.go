/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package indy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/schema"
	cursa "github.com/scoir/canis/pkg/ursa"
)

func (r *CredentialEngine) buildIndyCredential(issuerDID, schemaID, credDefID, offerNonce, blindedMasterSecret, blindedMSCorrectnessProof, requestNonce string,
	credDef *vdr.ClaimDefData, credDefPrivateKey string, values map[string]interface{}) (*decorator.AttachmentData, error) {

	blindedCredentialSecrets, err := ursa.BlindedCredentialSecretsFromJSON([]byte(blindedMasterSecret))
	if err != nil {
		return nil, err
	}

	blindedCredentialSecretsCorrectnessProof, err := ursa.BlindedCredentialSecretsCorrectnessProofFromJSON([]byte(blindedMSCorrectnessProof))
	if err != nil {
		return nil, err
	}

	credentialNonce, err := ursa.NonceFromJSON(fmt.Sprintf("\"%s\"", offerNonce))
	if err != nil {
		return nil, err
	}

	credentialIssuanceNonce, err := ursa.NonceFromJSON(requestNonce)
	if err != nil {
		return nil, err
	}

	encodedValues := schema.IndyCredentialValues{}

	builder, err := ursa.NewValueBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error from ursa value builder")
	}

	for k, v := range values {
		raw, enc := ursa.EncodeValue(v)

		err = builder.AddDecKnown(k, enc)
		if err != nil {
			return nil, errors.Wrap(err, "unexpected error adding to ursa value builder")
		}

		encodedValues[k] = &schema.IndyAttributeValue{Raw: raw, Encoded: enc}
	}

	credentialValues, err := builder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create values")
	}

	credentialPubKey, err := cursa.CredDefPublicKey(credDef.PKey(), credDef.RKey())
	if err != nil {
		return nil, err
	}

	credentialPrivKey, err := ursa.CredentialPrivateKeyFromJSON([]byte(credDefPrivateKey))
	if err != nil {
		return nil, err
	}

	signParams := ursa.NewSignatureParams()
	signParams.ProverID = issuerDID
	signParams.CredentialPubKey = credentialPubKey
	signParams.CredentialPrivKey = credentialPrivKey
	signParams.BlindedCredentialSecrets = blindedCredentialSecrets
	signParams.BlindedCredentialSecretsCorrectnessProof = blindedCredentialSecretsCorrectnessProof
	signParams.CredentialNonce = credentialNonce
	signParams.CredentialValues = credentialValues
	signParams.CredentialIssuanceNonce = credentialIssuanceNonce

	sig, sigCorrectnessProof, err := signParams.SignCredential()
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign credential")
	}

	defer func() {
		_ = blindedCredentialSecrets.Free()
		_ = blindedCredentialSecretsCorrectnessProof.Free()
		_ = credentialNonce.Free()
		_ = credentialIssuanceNonce.Free()
		_ = credentialValues.Free()
		_ = credentialPubKey.Free()
		_ = credentialPrivKey.Free()
	}()

	sigOut, err := sig.ToJSON()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get signature JSON")
	}

	proofOut, err := sigCorrectnessProof.ToJSON()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get correctness proof JSON")
	}
	cred := &schema.IndyCredential{
		SchemaID:                  schemaID,
		CredDefID:                 credDefID,
		Signature:                 sigOut,
		SignatureCorrectnessProof: proofOut,
		Values:                    encodedValues,
	}

	d, _ := json.Marshal(cred)
	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil
}

func (r *CredentialEngine) buildIndyOffer(schemaID, credDefID string, keyCorrectnessProof map[string]interface{}) (*decorator.AttachmentData, error) {
	nonce, err := r.oracle.NewNonce()
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error creating nonce")
	}

	offer := schema.IndyCredentialOffer{
		SchemaID:            schemaID,
		CredDefID:           credDefID,
		KeyCorrectnessProof: keyCorrectnessProof,
		Nonce:               strings.Trim(nonce, "\""),
	}

	d, _ := json.Marshal(offer)
	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil
}
