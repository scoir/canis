package engine

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
)

const PresentProofType = "https://didcomm.org/present-proof/2.0/request-presentation"

//go:generate mockery -name=PresentationEngine
type PresentationEngine interface {
	Accept(typ string) bool
	RequestPresentation(attrInfo map[string]*schema.IndyProofRequestAttr,
		predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error)
	RequestPresentationFormat() string
	Verify(presentation, request []byte, theirDID string, myDID string) error
}

//go:generate mockery -name=PresentationRegistry
type PresentationRegistry interface {
	RequestPresentation(typ string, attrInfo map[string]*schema.IndyProofRequestAttr,
		predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error)
	Verify(format string, presentation, request []byte, theirDID string, myDID string) error
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
func (r *Registry) RequestPresentation(typ string, attrInfo map[string]*schema.IndyProofRequestAttr,
	predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error) {

	e, err := r.resolveEngine(typ)
	if err != nil {
		return nil, err
	}

	return e.RequestPresentation(attrInfo, predicateInfo)
}

func (r *Registry) Verify(format string, presentation, request []byte, theirDID string, myDID string) error {

	e, err := r.resolveEngine(format)
	if err != nil {
		return err
	}

	return e.Verify(presentation, request, theirDID, myDID)

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
