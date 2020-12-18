package indy

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/schema"
)

func TestNew(t *testing.T) {
	t.Run("indy store error", func(t *testing.T) {
		prov := &providerMock{
			store: &storeMock{
				OpenStoreErr: errors.New("open store error"),
			},
		}

		engine, err := New(prov)
		require.Error(t, err)
		require.Nil(t, engine)
		require.Contains(t, err.Error(), "open store error")
	})

	t.Run("indy vdr error", func(t *testing.T) {
		prov := &providerMock{
			vdrError: errors.New("vdr error"),
			store: &storeMock{
				Store: &storage.MockStore{},
			},
		}

		engine, err := New(prov)
		require.Error(t, err)
		require.Nil(t, engine)
		require.Contains(t, err.Error(), "vdr error")
	})
}

func TestEngine_RequestPresentationAttach(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := &providerMock{
			store: &storeMock{
				Store: &storage.MockStore{},
			},
		}

		engine, err := New(prov)
		require.NoError(t, err)
		engine.crypto = &cryptoMock{}

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

		attach, err := engine.RequestPresentation(attrInfo, predInfo)
		require.NoError(t, err)
		require.NotNil(t, attach)

		expected := `eyJuYW1lIjoiUHJvb2YgbmFtZS4uLiIsInZlcnNpb24iOiIwLjAuMSIsIm5vbmNlIjoibm9uY2UiLCJyZXF1ZXN0ZWRfYXR0cmlidXRlcyI6eyJhdHRyMSI6eyJuYW1lIjoiYXR0ciBuYW1lIDEiLCJyZXN0cmljdGlvbnMiOiJyZXN0cmljdGlvbnMifX0sInJlcXVlc3RlZF9wcmVkaWNhdGVzIjp7InByZWQxIjp7Im5hbWUiOiJwcmVkaWNhdGUgbmFtZSAxIiwicF90eXBlIjoicHl0cGUiLCJwX3ZhbHVlIjoicHZhbHVlIiwicmVzdHJpY3Rpb25zIjoicmVzdHJpY3Rpb25zIn19fQ==`
		require.Equal(t, expected, attach)
	})

	t.Run("nonce error", func(t *testing.T) {
		prov := &providerMock{
			store: &storeMock{
				Store: &storage.MockStore{},
			},
		}

		engine, err := New(prov)
		require.NoError(t, err)
		engine.crypto = &cryptoMock{
			NewNonceErr: errors.New("nonce error"),
		}

		attach, err := engine.RequestPresentation(nil, nil)
		require.Error(t, err)
		require.Empty(t, attach)
		require.Contains(t, err.Error(), "nonce error")

	})
}

func TestUrsaCrypto_NewNonce(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		prov := &providerMock{
			store: &storeMock{
				Store: &storage.MockStore{},
			},
		}

		engine, err := New(prov)
		require.NoError(t, err)
		nonce, err := engine.crypto.NewNonce()

		require.NoError(t, err)
		require.NotEmpty(t, nonce)
	})
}

func TestEngine_Accept(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := &providerMock{
			store: &storeMock{
				Store: &storage.MockStore{},
			},
		}

		engine, err := New(prov)
		require.NoError(t, err)
		require.True(t, engine.Accept(Format))
		require.False(t, engine.Accept("unknown"))
	})
}

func TestEngine_RequestPresentationFormat(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := &providerMock{
			store: &storeMock{
				Store: &storage.MockStore{},
			},
		}

		engine, err := New(prov)
		require.NoError(t, err)
		require.Equal(t, Format, engine.RequestPresentationFormat())
	})
}
