package indy

import (
	"crypto/ed25519"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"

	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
)

func TestVDRI_Build(t *testing.T) {
	t.Run("illegal key type", func(t *testing.T) {
		r := &VDRI{
			methodName: "sov",
		}

		pubKey := &vdriapi.PubKey{
			ID:    "test",
			Value: "test",
			Type:  "Not Valid",
		}

		doc, err := r.Build(pubKey)
		require.Nil(t, doc)
		require.Error(t, err)
	})

	t.Run("valid key", func(t *testing.T) {
		r := &VDRI{
			methodName: "sov",
		}

		k := ed25519.NewKeyFromSeed([]byte("b2352b32947e188eb72871093ac6217e"))
		pubKey := &vdriapi.PubKey{
			ID:    "test",
			Value: base58.Encode(k),
			Type:  "Ed25519VerificationKey2018",
		}

		doc, err := r.Build(pubKey)
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Equal(t, "did:sov:D8HmB7s9KCGuPGbi5Ymiqr", doc.ID)
		require.NotNil(t, doc.Context)
		require.NotNil(t, doc.Updated)
		require.NotNil(t, doc.Created)
		require.Len(t, doc.Authentication, 1)
		require.Len(t, doc.PublicKey, 1)
		require.Nil(t, doc.Service)
	})

	t.Run("valid key with service endpoint", func(t *testing.T) {
		r := &VDRI{
			methodName: "sov",
		}

		k := ed25519.NewKeyFromSeed([]byte("b2352b32947e188eb72871093ac6217e"))
		pubKey := &vdriapi.PubKey{
			ID:    "test",
			Value: base58.Encode(k),
			Type:  "Ed25519VerificationKey2018",
		}

		ep := "http://127.0.0.1:8080"
		doc, err := r.Build(pubKey, vdriapi.WithServiceType(vdriapi.DIDCommServiceType), vdriapi.WithServiceEndpoint(ep))
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Equal(t, "did:sov:D8HmB7s9KCGuPGbi5Ymiqr", doc.ID)
		require.NotNil(t, doc.Context)
		require.NotNil(t, doc.Updated)
		require.NotNil(t, doc.Created)
		require.Len(t, doc.Authentication, 1)
		require.Len(t, doc.PublicKey, 1)

		require.NotNil(t, doc.Service)
		require.Len(t, doc.Service, 1)
		require.Equal(t, ep, doc.Service[0].ServiceEndpoint)
	})
}
