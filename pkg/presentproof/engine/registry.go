package engine

import (
	"fmt"

	"github.com/scoir/canis/pkg/datastore"
)

//go:generate mockery -name=PresentationEngine
type PresentationEngine interface {
	Accept(typ string) bool
}

//go:generate mockery -name=PresentationRegistry
type PresentationRegistry interface {
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
