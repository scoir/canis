package verifier

import (
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore"
	dsmocks "github.com/scoir/canis/pkg/datastore/mocks"
	"github.com/scoir/canis/pkg/presentproof/engine/mocks"
)

type mockProps struct {
	myDID    string
	theirDID string
	piid     string
}

func (r *mockProps) All() map[string]interface{} { return nil }

func (r *mockProps) MyDID() string {
	return r.myDID
}
func (r *mockProps) TheirDID() string {
	return r.theirDID
}
func (r *mockProps) PIID() string {
	return r.piid
}

type badProps struct {
}

func (r *badProps) All() map[string]interface{} { return nil }

func TestProofHandler_PresentationMsg(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		var result string
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Continue: func(p interface{}) {
				result = p.(string)
			},
			Properties: &mockProps{piid: "123", theirDID: "sov:123", myDID: "sov:abc"},
		}

		pr := &datastore.PresentationRequest{
			Data: []byte(`proofData`),
		}

		p := &ppprotocol.Presentation{
			Formats: []ppprotocol.Format{{
				AttachID: "abc",
				Format:   "indy",
			}},
			PresentationsAttach: []decorator.Attachment{{
				ID: "abc",
				Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				},
			}},
		}

		verified := &datastore.Presentation{
			TheirDID: "sov:123",
			MyDID:    "sov:abc",
			Format:   "indy",
			Data:     []byte(`{}`),
		}

		suite.store.On("GetPresentationRequest", "123").Return(pr, nil)
		suite.registry.On("Verify", "indy", []byte(`{}`), []byte(`proofData`), "sov:123", "sov:abc").Return(nil)
		suite.store.On("InsertPresentation", verified).Return("id-1", nil)

		suite.target.PresentationMsg(action, p)
		require.NoError(t, err)
		require.Equal(t, "123", result)
	})
	t.Run("unable to store presentation", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Properties: &mockProps{piid: "123", theirDID: "sov:123", myDID: "sov:abc"},
		}

		pr := &datastore.PresentationRequest{
			Data: []byte(`proofData`),
		}

		p := &ppprotocol.Presentation{
			Formats: []ppprotocol.Format{{
				AttachID: "abc",
				Format:   "indy",
			}},
			PresentationsAttach: []decorator.Attachment{{
				ID: "abc",
				Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				},
			}},
		}

		verified := &datastore.Presentation{
			TheirDID: "sov:123",
			MyDID:    "sov:abc",
			Format:   "indy",
			Data:     []byte(`{}`),
		}

		suite.store.On("GetPresentationRequest", "123").Return(pr, nil)
		suite.registry.On("Verify", "indy", []byte(`{}`), []byte(`proofData`), "sov:123", "sov:abc").Return(nil)
		suite.store.On("InsertPresentation", verified).Return("", errors.New("not saved"))

		suite.target.PresentationMsg(action, p)
		require.Error(t, err)
		require.Equal(t, "unexpected error saving verified presention: (not saved)", err.Error())

	})
	t.Run("registry error", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Properties: &mockProps{piid: "123", theirDID: "sov:123", myDID: "sov:abc"},
		}

		pr := &datastore.PresentationRequest{
			Data: []byte(`proofData`),
		}

		p := &ppprotocol.Presentation{
			Formats: []ppprotocol.Format{{
				AttachID: "abc",
				Format:   "indy",
			}},
			PresentationsAttach: []decorator.Attachment{{
				ID: "abc",
				Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				},
			}},
		}

		suite.store.On("GetPresentationRequest", "123").Return(pr, nil)
		suite.registry.On("Verify", "indy", []byte(`{}`), []byte(`proofData`), "sov:123", "sov:abc").Return(errors.New("boom"))

		suite.target.PresentationMsg(action, p)
		require.Error(t, err)
		require.Equal(t, "unexpected error verifying 0 presentation: (boom)", err.Error())

	})
	t.Run("attachment mismatch", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Properties: &mockProps{piid: "123"},
		}

		pr := &datastore.PresentationRequest{}

		p := &ppprotocol.Presentation{
			Formats: []ppprotocol.Format{{
				AttachID: "xyz",
				Format:   "indy",
			}},
			PresentationsAttach: []decorator.Attachment{{
				ID: "abc",
				Data: decorator.AttachmentData{
					JSON: map[string]interface{}{},
				},
			}},
		}

		suite.store.On("GetPresentationRequest", "123").Return(pr, nil)

		suite.target.PresentationMsg(action, p)
		require.Error(t, err)
		require.Equal(t, "presentations and formats do not match 0", err.Error())

	})
	t.Run("empty attachment", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Properties: &mockProps{piid: "123"},
		}

		pr := &datastore.PresentationRequest{}

		p := &ppprotocol.Presentation{
			Formats:             []ppprotocol.Format{{}},
			PresentationsAttach: []decorator.Attachment{{}},
		}

		suite.store.On("GetPresentationRequest", "123").Return(pr, nil)

		suite.target.PresentationMsg(action, p)
		require.Error(t, err)
		require.Equal(t, "unable to fetch presentation data from proof 0: (no contents in this attachment)", err.Error())

	})
	t.Run("bad piid", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Properties: &mockProps{piid: "123"},
		}

		p := &ppprotocol.Presentation{}

		suite.store.On("GetPresentationRequest", "123").Return(nil, errors.New("not found"))

		suite.target.PresentationMsg(action, p)
		require.Error(t, err)
		require.Equal(t, "unable to find presentation request 123", err.Error())

	})
	t.Run("invalid props", func(t *testing.T) {
		suite, cleanup := setup(t)
		defer cleanup()

		var err error
		action := service.DIDCommAction{
			ProtocolName: "presentation",
			Stop: func(e error) {
				err = e
			},
			Properties: &badProps{},
		}

		p := &ppprotocol.Presentation{}

		suite.target.PresentationMsg(action, p)
		require.Error(t, err)
		require.Equal(t, "presentation properties invalid", err.Error())

	})
}

func TestProofHandler_PresentationPreviewMsg(t *testing.T) {
	suite, cleanup := setup(t)
	defer cleanup()

	var err error
	action := service.DIDCommAction{
		ProtocolName: "presentation-preview",
		Stop: func(e error) {
			err = e
		},
	}

	suite.target.PresentationPreviewMsg(action, nil)
	require.Error(t, err)
	require.Equal(t, "presentation preview not implemented", err.Error())
}

func TestProofHandler_ProposePresentationMsg(t *testing.T) {
	suite, cleanup := setup(t)
	defer cleanup()

	var err error
	action := service.DIDCommAction{
		ProtocolName: "propose-presentation",
		Stop: func(e error) {
			err = e
		},
	}

	suite.target.ProposePresentationMsg(action, nil)
	require.Error(t, err)
	require.Equal(t, "presentation proposal not implemented", err.Error())
}

func TestProofHandler_RequestPresentationMsg(t *testing.T) {
	suite, cleanup := setup(t)
	defer cleanup()

	var err error
	action := service.DIDCommAction{
		ProtocolName: "request-presentation",
		Stop: func(e error) {
			err = e
		},
	}

	suite.target.RequestPresentationMsg(action, nil)
	require.Error(t, err)
	require.Equal(t, "request presentation not implemented", err.Error())
}

type suite struct {
	target   *proofHandler
	store    *dsmocks.Store
	registry *mocks.PresentationRegistry
}

func setup(t *testing.T) (*suite, func()) {
	s := &suite{
		store:    &dsmocks.Store{},
		registry: &mocks.PresentationRegistry{},
	}

	s.target = &proofHandler{
		store:    s.store,
		registry: s.registry,
	}

	return s, func() {
		s.store.AssertExpectations(t)
		s.registry.AssertExpectations(t)
	}
}
