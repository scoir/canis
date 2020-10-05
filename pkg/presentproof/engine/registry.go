package engine

import (
	"fmt"
	"github.com/google/uuid"
	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/verifier/api"
)

const PresentProofType = "https://didcomm.org/present-proof/2.0/request-presentation"

//go:generate mockery -name=PresentationEngine
type PresentationEngine interface {
	RequestPresentationAttach(attrInfo map[string]*api.AttrInfo, predicateInfo map[string]*api.PredicateInfo) ([]byte, error)
	RequestPresentationFormat() string
	Accept(typ string) bool
}

//go:generate mockery -name=PresentationRegistry
type PresentationRegistry interface {
	RequestPresentation(typ string, attrInfo map[string]*api.AttrInfo, predicateInfo map[string]*api.PredicateInfo) (
		*ppclient.RequestPresentation, error)
}

type Option func(opts *Registry)

type Registry struct {
	engines  []PresentationEngine
	didStore datastore.Store
}

type provider interface {
	Store() datastore.Store
}

func New(prov provider, opts ...Option) *Registry {
	reg := &Registry{didStore: prov.Store(), engines: []PresentationEngine{}}

	for _, opt := range opts {
		opt(reg)
	}

	return reg
}

// RequestPresentation
func (r *Registry) RequestPresentation(typ string, attrInfo map[string]*api.AttrInfo,
	predicateInfo map[string]*api.PredicateInfo) (*ppclient.RequestPresentation, error) {

	e, err := r.resolveEngine(typ)
	if err != nil {
		return nil, err
	}

	attach, err := e.RequestPresentationAttach(attrInfo, predicateInfo)
	if err != nil {
		return nil, err
	}

	attachID := uuid.New().String()
	req := &ppclient.RequestPresentation{
		Type: PresentProofType,
		Formats: []presentproof.Format{{
			AttachID: attachID,
			Format:   e.RequestPresentationFormat(),
		}},
		RequestPresentationsAttach: []decorator.Attachment{
			{
				ID:       attachID,
				MimeType: "application/json",
				Data: decorator.AttachmentData{
					Base64: string(attach),
				},
			},
		},
	}

	return req, nil
}

func (r *Registry) resolveEngine(method string) (PresentationEngine, error) {
	for _, e := range r.engines {
		if e.Accept(method) {
			return e, nil
		}
	}

	return nil, fmt.Errorf("presentation type %s not supported by any engine", method)
}

// WithEngine adds did method implementation for store.
func WithEngine(e PresentationEngine) Option {
	return func(opts *Registry) {
		opts.engines = append(opts.engines, e)
	}
}
