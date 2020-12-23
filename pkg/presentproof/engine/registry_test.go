package engine

import (
	"encoding/base64"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
	"github.com/scoir/canis/pkg/schema"
)

type providerMock struct {
	store *mocks.Store
}

func (r *providerMock) Store() datastore.Store {
	return r.store
}

func NewProvider() *providerMock {
	return &providerMock{store: &mocks.Store{}}
}

func TestRegistry_RequestPresentation(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&engineMock{
			Format:     "format",
			DoesAccept: true,
			RequestPresentationAttachFunc: func(attrInfo map[string]*schema.IndyProofRequestAttr, predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error) {
				return &decorator.AttachmentData{Base64: base64.StdEncoding.EncodeToString([]byte("test"))}, nil
			},
			RequestPresentationAttachErr: nil,
		}))

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

		presentation, err := reg.RequestPresentation("name", "version", "t", attrInfo, predInfo)
		require.NoError(t, err)
		require.NotNil(t, presentation)

		require.NoError(t, err)
		require.Equal(t, "dGVzdA==", presentation.Base64)
	})

	t.Run("engine error", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&engineMock{
			DoesAccept: false,
		}))

		presentation, err := reg.RequestPresentation("name", "version", "t", nil, nil)
		require.Error(t, err)
		require.Nil(t, presentation)
		require.Contains(t, err.Error(), "presentation type t not supported by any engine")
	})

	t.Run("RequestPresentation error", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&engineMock{
			Format:                       "format",
			DoesAccept:                   true,
			RequestPresentationAttachErr: errors.New("RequestPresentation error"),
		}))

		presentation, err := reg.RequestPresentation("name", "version", "t", nil, nil)
		require.Error(t, err)
		require.Nil(t, presentation)
		require.Contains(t, err.Error(), "RequestPresentation error")
	})
}

func TestVerify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&engineMock{
			Format:     "format",
			DoesAccept: true,
		}))

		err := reg.Verify("format", []byte{}, []byte{}, "did:sov:123", "did:sov:abc")
		require.NoError(t, err)
	})
	t.Run("no engine error", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&engineMock{
			Format:     "format",
			DoesAccept: false,
		}))

		err := reg.Verify("format", []byte{}, []byte{}, "did:sov:123", "did:sov:abc")
		require.Error(t, err)
	})
}

type engineMock struct {
	Format                        string
	DoesAccept                    bool
	RequestPresentationAttachFunc func(attrInfo map[string]*schema.IndyProofRequestAttr, predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error)
	RequestPresentationAttachErr  error
	VerifyErr                     error
}

func (r *engineMock) RequestPresentation(name, version string, attrInfo map[string]*schema.IndyProofRequestAttr, predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error) {
	if r.RequestPresentationAttachErr != nil {
		return nil, r.RequestPresentationAttachErr
	}

	if r.RequestPresentationAttachFunc != nil {
		return r.RequestPresentationAttachFunc(attrInfo, predicateInfo)
	}

	return nil, nil
}

func (r *engineMock) Verify(presentation, request []byte, theirDID string, myDID string) error {
	return r.VerifyErr
}

func (r *engineMock) Accept(_ string) bool {
	return r.DoesAccept
}

func (r *engineMock) RequestPresentationFormat() string {
	return r.Format
}
