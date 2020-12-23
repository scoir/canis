package indy

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

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
	t.Run("test", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		crypto := &schema.CryptoProof{
			Proofs: []*schema.SubProof{
				{},
			},
		}

		cryptoData, err := json.Marshal(crypto)
		require.NoError(t, err)

		proof := &schema.IndyProof{
			Proof: cryptoData,
			RequestedProof: &schema.IndyRequestedProof{
				RevealedAttrs: map[string]*schema.RevealedAttributeInfo{
					"attr1": {
						SubProofIndex: 0,
						Raw:           "test",
						Encoded:       "12345",
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
			Nonce:   "1234567890987654321",
			RequestedAttributes: map[string]*schema.IndyProofRequestAttr{"attr1": {
				Name: "attr1",
			}},
			RequestedPredicates: map[string]*schema.IndyProofRequestPredicate{},
		}
		reqData, err := json.Marshal(proofRequest)
		require.NoError(t, err)

		err = engine.Verify(proofData, reqData, "did:sov:123", "did:sov:abc")
		require.NoError(t, err)

		prov.Asserts(t)
	})
}

type provider struct {
	provider *MockProvider
	oracle   *MockOracle
	vdr      *MockVDRClient
	store    *dsmocks.Store
}

func newProvider() provider {
	p := provider{
		provider: &MockProvider{},
		oracle:   &MockOracle{},
		vdr:      &MockVDRClient{},
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
