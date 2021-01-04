package indy

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/stretchr/testify/require"

	indymocks "github.com/scoir/canis/pkg/credential/engine/indy/mocks"
	"github.com/scoir/canis/pkg/datastore"
	dsmocks "github.com/scoir/canis/pkg/datastore/mocks"
	"github.com/scoir/canis/pkg/schema"
)

func TestNew(t *testing.T) {
	t.Run("indy vdr error", func(t *testing.T) {
		prov := provider{
			provider: &MockProvider{},
		}

		prov.provider.On("IndyVDR").Return(nil, errors.New("vdr error"))

		engine, err := New(prov.provider)
		require.Error(t, err)
		require.Nil(t, engine)
		require.Contains(t, err.Error(), "vdr error")

		prov.provider.AssertExpectations(t)
	})
}

func TestEngine_RequestPresentationAttach(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		attrInfo := make(map[string]*schema.IndyProofRequestAttr)
		predInfo := make(map[string]*schema.IndyProofRequestPredicate)

		attrInfo["attr1"] = &schema.IndyProofRequestAttr{
			Name:         "attr name 1",
			Restrictions: "restrictions",
		}

		predInfo["pred1"] = &schema.IndyProofRequestPredicate{
			Name:         "predicate name 1",
			PType:        "pytpe",
			PValue:       32,
			Restrictions: "restrictions",
		}

		prov.oracle.On("NewNonce").Return("1234567890987654321", nil)

		attach, err := engine.RequestPresentation("name", "ver", attrInfo, predInfo)
		require.NoError(t, err)
		require.NotNil(t, attach)

		expected := `eyJuYW1lIjoibmFtZSIsInZlcnNpb24iOiJ2ZXIiLCJub25jZSI6IjEyMzQ1Njc4OTA5ODc2NTQzMjEiLCJyZXF1ZXN0ZWRfYXR0cmlidXRlcyI6eyJhdHRyMSI6eyJuYW1lIjoiYXR0ciBuYW1lIDEiLCJuYW1lcyI6bnVsbCwicmVzdHJpY3Rpb25zIjoicmVzdHJpY3Rpb25zIiwibm9uX3Jldm9rZWQiOnsiZnJvbSI6MCwidG8iOjB9fX0sInJlcXVlc3RlZF9wcmVkaWNhdGVzIjp7InByZWQxIjp7Im5hbWUiOiJwcmVkaWNhdGUgbmFtZSAxIiwicF90eXBlIjoicHl0cGUiLCJwX3ZhbHVlIjozMiwicmVzdHJpY3Rpb25zIjoicmVzdHJpY3Rpb25zIiwibm9uX3Jldm9rZWQiOnsiZnJvbSI6MCwidG8iOjB9fX0sIm5vbl9yZXZva2VkIjp7ImZyb20iOjAsInRvIjowfX0=`

		require.Equal(t, expected, attach.Base64)

		prov.Asserts(t)

	})

	t.Run("nonce error", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		prov.oracle.On("NewNonce").Return("", errors.New("nonce error"))

		attach, err := engine.RequestPresentation("", "", nil, nil)
		require.Error(t, err)
		require.Empty(t, attach)
		require.Contains(t, err.Error(), "nonce error")

		prov.Asserts(t)

	})
}

func TestEngine_Accept(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)
		require.True(t, engine.Accept(Format))
		require.False(t, engine.Accept("unknown"))

		prov.Asserts(t)
	})
}

func TestEngine_RequestPresentationFormat(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)
		require.Equal(t, Format, engine.RequestPresentationFormat())

		prov.Asserts(t)
	})
}

func TestEngine_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)

		s := &datastore.Schema{
			ID:               "schema-1",
			ExternalSchemaID: "123:2:cl:foo",
			Attributes: []*datastore.Attribute{
				{Name: "attr1"},
			},
		}

		cryptoProof := getProof(t)
		proof := &schema.IndyProof{
			Proof: cryptoProof.proofJSON,
			RequestedProof: &schema.IndyRequestedProof{
				RevealedAttrs: map[string]*schema.RevealedAttributeInfo{
					"attr1": {
						SubProofIndex: 0,
						Raw:           "test-val-1",
						Encoded:       "19784575220953737155389574731041019715457831279719061977145029012931741524053",
					},
				},
				UnrevealedAttrs: map[string]*schema.SubProofReferent{
					"master_secret": {
						SubProofIndex: 0,
					},
				},
			},
			Identifiers: []*schema.Identifier{
				{
					SchemaID:  "123:2:cl:foo",
					CredDefID: "abc:3:cl:bar",
				},
			},
		}
		proofData, err := json.Marshal(proof)
		require.NoError(t, err)

		proofRequest := &PresentationRequest{
			Name:    "name",
			Version: "0.0.0",
			Nonce:   string(cryptoProof.credentialNonce),
			RequestedAttributes: map[string]*schema.IndyProofRequestAttr{
				"attr1": {
					Name: "attr1",
				},
				"master_secret": {
					Name: "master_secret",
				},
			},
			RequestedPredicates: map[string]*schema.IndyProofRequestPredicate{},
		}
		reqData, err := json.Marshal(proofRequest)
		require.NoError(t, err)

		pubKey := map[string]interface{}{}
		err = json.Unmarshal(cryptoProof.credDefPubKeyJSON, &pubKey)
		require.NoError(t, err)
		rply := &vdr.ReadReply{Data: map[string]interface{}{
			"primary": pubKey["p_key"],
		}}

		prov.vdr.On("GetCredDef", "abc:3:cl:bar").Return(rply, nil)
		prov.store.On("GetSchemaByExternalID", "123:2:cl:foo").Return(s, nil)

		err = engine.Verify(proofData, reqData, "did:sov:123", "did:sov:abc")
		require.NoError(t, err)

	})
}

type provider struct {
	provider *MockProvider
	oracle   *indymocks.Oracle
	vdr      *indymocks.VDRClient
	store    *dsmocks.Store
}

func newProvider() provider {
	p := provider{
		provider: &MockProvider{},
		oracle:   &indymocks.Oracle{},
		vdr:      &indymocks.VDRClient{},
		store:    &dsmocks.Store{},
	}

	p.provider.On("IndyVDR").Return(p.vdr, nil)
	p.provider.On("Oracle").Return(p.oracle)
	p.provider.On("Store").Return(p.store)

	return p
}

func (r *provider) Asserts(t *testing.T) {
	r.provider.AssertExpectations(t)
	r.oracle.AssertExpectations(t)
	r.vdr.AssertExpectations(t)
	r.store.AssertExpectations(t)
}

type proofData struct {
	credDefPubKeyJSON []byte
	proofJSON         []byte
	credentialNonce   []byte
	masterSecret      string
}

func getProof(t *testing.T) proofData {
	out := proofData{}

	schemaBuilder, err := ursa.NewCredentialSchemaBuilder()
	require.NoError(t, err)
	err = schemaBuilder.AddAttr("attr1")
	require.NoError(t, err)
	sch, err := schemaBuilder.Finalize()
	require.NoError(t, err)

	nonSchemaBuilder, err := ursa.NewNonCredentialSchemaBuilder()
	require.NoError(t, err)
	err = nonSchemaBuilder.AddAttr("master_secret")
	require.NoError(t, err)
	nonSchema, err := nonSchemaBuilder.Finalize()

	credDef, err := ursa.NewCredentialDef(sch, nonSchema, false)
	require.NoError(t, err)
	require.NotNil(t, credDef)

	js, err := credDef.PubKey.ToJSON()
	require.NoError(t, err)
	out.credDefPubKeyJSON = js

	masterSecret, err := ursa.NewMasterSecret()
	require.NoError(t, err)
	js, err = masterSecret.ToJSON()
	require.NoError(t, err)
	m := struct {
		MS string `json:"ms"`
	}{}
	err = json.Unmarshal(js, &m)
	require.NoError(t, err)

	out.masterSecret = m.MS

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

	blindedSecrets, err := ursa.BlindCredentialSecrets(credDef.PubKey, credDef.KeyCorrectnessProof, credentialNonce, values)
	require.NoError(t, err)

	js, err = blindedSecrets.Handle.ToJSON()
	require.NoError(t, err)

	js, err = blindedSecrets.CorrectnessProof.ToJSON()
	require.NoError(t, err)

	credentialIssuanceNonce, err := ursa.NewNonce()
	require.NoError(t, err)

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
	require.NoError(t, err)

	err = credSig.ProcessCredentialSignature(values, credSigKP, blindedSecrets.BlindingFactor, credDef.PubKey, credentialIssuanceNonce)
	require.NoError(t, err)

	subProofBuilder, err := ursa.NewSubProofRequestBuilder()
	require.NoError(t, err)
	err = subProofBuilder.AddRevealedAttr("attr1")
	require.NoError(t, err)
	subProofRequest, err := subProofBuilder.Finalize()
	require.NoError(t, err)

	proofBuilder, err := ursa.NewProofBuilder()
	require.NoError(t, err)
	err = proofBuilder.AddCommonAttribute("master_secret")
	require.NoError(t, err)
	err = proofBuilder.AddSubProofRequest(subProofRequest, sch, nonSchema, credSig, values, credDef.PubKey)
	require.NoError(t, err)

	proofRequestNonce, err := ursa.NewNonce()
	require.NoError(t, err)

	js, err = proofRequestNonce.ToJSON()
	require.NoError(t, err)
	out.credentialNonce = js

	proof, err := proofBuilder.Finalize(proofRequestNonce)
	require.NoError(t, err)

	js, err = proof.ToJSON()
	require.NoError(t, err)

	out.proofJSON = js

	verifier, err := ursa.NewProofVerifier()
	require.NoError(t, err)

	err = verifier.AddSubProofRequest(subProofRequest, sch, nonSchema, credDef.PubKey)
	require.NoError(t, err)

	err = verifier.Verify(proof, proofRequestNonce)
	require.NoError(t, err)

	return out
}
