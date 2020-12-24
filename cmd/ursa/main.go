package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func main() {
	t := &testing.T{}
	schemaBuilder, err := ursa.NewCredentialSchemaBuilder()
	require.NoError(t, err)
	err = schemaBuilder.AddAttr("attr1")
	require.NoError(t, err)
	schema, err := schemaBuilder.Finalize()
	require.NoError(t, err)

	nonSchemaBuilder, err := ursa.NewNonCredentialSchemaBuilder()
	require.NoError(t, err)
	err = nonSchemaBuilder.AddAttr("master_secret")
	require.NoError(t, err)
	nonSchema, err := nonSchemaBuilder.Finalize()

	credDef, err := ursa.NewCredentialDef(schema, nonSchema, false)
	require.NoError(t, err)
	require.NotNil(t, credDef)

	js, err := credDef.PubKey.ToJSON()
	require.NoError(t, err)
	//fmt.Println(string(js))

	masterSecret, err := ursa.NewMasterSecret()
	require.NoError(t, err)
	js, err = masterSecret.ToJSON()
	require.NoError(t, err)
	m := struct {
		MS string `json:"ms"`
	}{}
	err = json.Unmarshal(js, &m)
	require.NoError(t, err)

	//fmt.Println(m.MS)

	valuesBuilder, err := ursa.NewValueBuilder()
	require.NoError(t, err)
	err = valuesBuilder.AddDecHidden("master_secret", m.MS)
	require.NoError(t, err)

	_, enc := ursa.EncodeValue("test-val-1")
	fmt.Println(enc)

	err = valuesBuilder.AddDecKnown("attr1", enc)
	require.NoError(t, err)

	values, err := valuesBuilder.Finalize()
	require.NoError(t, err)

	credentialNonce, err := ursa.NewNonce()
	require.NoError(t, err)

	js, err = credentialNonce.ToJSON()
	require.NoError(t, err)
	//fmt.Println(string(js))

	blindedSecrets, err := ursa.BlindCredentialSecrets(credDef.PubKey, credDef.KeyCorrectnessProof, credentialNonce, values)
	require.NoError(t, err)

	js, err = blindedSecrets.Handle.ToJSON()
	require.NoError(t, err)
	//fmt.Println(string(js))
	js, err = blindedSecrets.CorrectnessProof.ToJSON()
	require.NoError(t, err)
	//fmt.Println(string(js))

	credentialIssuanceNonce, err := ursa.NewNonce()
	assert.NoError(t, err)

	p := ursa.SignatureParams{
		ProverID:                                 "CnEDk9HrMnmiHXEV1WFgbVCRteYnPqsJwrTdcZaNhFVW",
		BlindedCredentialSecrets:                 blindedSecrets.Handle,
		BlindedCredentialSecretsCorrectnessProof: blindedSecrets.CorrectnessProof,
		CredentialIssuanceNonce:                  credentialIssuanceNonce,
		CredentialNonce:                          credentialNonce,
		CredentialValues:                         values,
		CredentialPubKey:                         credDef.PubKey,
		CredentialPrivKey:                        credDef.PrivKey,
	}

	credSig, credSigKP, err := p.SignCredential()
	assert.NoError(t, err)

	err = credSig.ProcessCredentialSignature(values, credSigKP, blindedSecrets.BlindingFactor, credDef.PubKey, credentialIssuanceNonce)
	assert.NoError(t, err)

	subProofBuilder, err := ursa.NewSubProofRequestBuilder()
	assert.NoError(t, err)
	err = subProofBuilder.AddRevealedAttr("attr1")
	assert.NoError(t, err)
	subProofRequest, err := subProofBuilder.Finalize()
	assert.NoError(t, err)

	proofBuilder, err := ursa.NewProofBuilder()
	assert.NoError(t, err)
	err = proofBuilder.AddCommonAttribute("master_secret")
	assert.NoError(t, err)
	err = proofBuilder.AddSubProofRequest(subProofRequest, schema, nonSchema, credSig, values, credDef.PubKey)
	assert.NoError(t, err)

	proofRequestNonce, err := ursa.NewNonce()
	assert.NoError(t, err)

	proof, err := proofBuilder.Finalize(proofRequestNonce)
	assert.NoError(t, err)

	verifier, err := ursa.NewProofVerifier()
	assert.NoError(t, err)

	err = verifier.AddSubProofRequest(subProofRequest, schema, nonSchema, credDef.PubKey)
	assert.NoError(t, err)

	err = verifier.Verify(proof, proofRequestNonce)
	assert.NoError(t, err)
}
