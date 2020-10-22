package engine

import (
	"encoding/base64"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
	api "github.com/scoir/canis/pkg/protogen/common"
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

		reg := New(prov, WithEngine(&indyProofMock{
			Format:     "format",
			DoesAccept: true,
			RequestPresentationAttachFunc: func(attrInfo map[string]*api.AttrInfo, predicateInfo map[string]*api.PredicateInfo) (string, error) {
				return base64.StdEncoding.EncodeToString([]byte("test")), nil
			},
			RequestPresentationAttachErr: nil,
		}))

		attrInfo := make(map[string]*api.AttrInfo)
		predInfo := make(map[string]*api.PredicateInfo)

		attrInfo["attr1"] = &api.AttrInfo{
			Name:         "attr name 1",
			Restrictions: "restrictions",
			NonRevoked:   nil,
		}

		predInfo["pred1"] = &api.PredicateInfo{
			Name:         "predicate name 1",
			PType:        "pytpe",
			PValue:       "pvalue",
			Restrictions: "restrictions",
			NonRevoked:   nil,
		}

		presentation, err := reg.RequestPresentation("t", attrInfo, predInfo)
		require.NoError(t, err)
		require.NotNil(t, presentation)

		require.Equal(t, PresentProofType, presentation.Type)
		require.Len(t, presentation.Formats, 1)
		require.Len(t, presentation.RequestPresentationsAttach, 1)

		require.NotEmpty(t, presentation.Formats[0].AttachID)
		require.Equal(t, "format", presentation.Formats[0].Format)

		require.NotEmpty(t, presentation.RequestPresentationsAttach[0].ID)
		require.Equal(t, "application/json", presentation.RequestPresentationsAttach[0].MimeType)

		require.NoError(t, err)
		require.Equal(t, "dGVzdA==", presentation.RequestPresentationsAttach[0].Data.Base64)
	})

	t.Run("engine error", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&indyProofMock{
			DoesAccept: false,
		}))

		presentation, err := reg.RequestPresentation("t", nil, nil)
		require.Error(t, err)
		require.Nil(t, presentation)
		require.Contains(t, err.Error(), "presentation type t not supported by any engine")
	})

	t.Run("RequestPresentationAttach error", func(t *testing.T) {
		prov := NewProvider()

		reg := New(prov, WithEngine(&indyProofMock{
			Format:                       "format",
			DoesAccept:                   true,
			RequestPresentationAttachErr: errors.New("RequestPresentationAttach error"),
		}))

		presentation, err := reg.RequestPresentation("t", nil, nil)
		require.Error(t, err)
		require.Nil(t, presentation)
		require.Contains(t, err.Error(), "RequestPresentationAttach error")
	})
}

type indyProofMock struct {
	Format                        string
	DoesAccept                    bool
	RequestPresentationAttachFunc func(attrInfo map[string]*api.AttrInfo, predicateInfo map[string]*api.PredicateInfo) (string, error)
	RequestPresentationAttachErr  error
}

func (r *indyProofMock) RequestPresentationAttach(attrInfo map[string]*api.AttrInfo,
	predicateInfo map[string]*api.PredicateInfo) (string, error) {
	if r.RequestPresentationAttachErr != nil {
		return "", r.RequestPresentationAttachErr
	}

	if r.RequestPresentationAttachFunc != nil {
		return r.RequestPresentationAttachFunc(attrInfo, predicateInfo)
	}

	return "", nil
}

func (r *indyProofMock) Accept(_ string) bool {
	return r.DoesAccept
}

func (r *indyProofMock) RequestPresentationFormat() string {
	return r.Format
}
