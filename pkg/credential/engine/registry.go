package engine

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/ursa"
)

type CredentialEngine interface {
	Accept(typ string) bool
	CreateSchema(issuer *datastore.DID, s *datastore.Schema) (string, error)
	RegisterSchema(registrant *datastore.DID, s *datastore.Schema) error
	CreateCredentialOffer(issuer *datastore.DID, s *datastore.Schema) (*decorator.AttachmentData, error)
	IssueCredential(s *datastore.Schema, offer *ursa.CredentialOffer, request *ursa.CredentialRequest, values *ursa.CredentialValues) (*decorator.AttachmentData, error)
}

//go:generate mockery -name=CredentialRegistry
type CredentialRegistry interface {
	CreateSchema(s *datastore.Schema) (string, error)
	RegisterSchema(registrant *datastore.DID, s *datastore.Schema) error
	CreateCredentialOffer(issuer *datastore.DID, s *datastore.Schema) (*decorator.AttachmentData, error)
	IssueCredential(issuer *datastore.DID, s *datastore.Schema) (*decorator.AttachmentData, error)
}

type Option func(opts *Registry)

type Registry struct {
	engines  []CredentialEngine
	didStore datastore.Store
}

type provider interface {
	Store() datastore.Store
}

func New(prov provider, opts ...Option) *Registry {
	reg := &Registry{didStore: prov.Store(), engines: []CredentialEngine{}}

	for _, opt := range opts {
		opt(reg)
	}

	return reg
}

func (r *Registry) CreateSchema(s *datastore.Schema) (string, error) {
	e, err := r.resolveEngine(s.Type)
	if err != nil {
		return "", err
	}

	issuer, err := r.didStore.GetPublicDID()
	if err != nil {
		return "", errors.Wrap(err, "error getting public did to create schema")
	}

	id, err := e.CreateSchema(issuer, s)
	return id, errors.Wrap(err, "error from credential engine")
}

func (r *Registry) RegisterSchema(registrant *datastore.DID, s *datastore.Schema) error {
	e, err := r.resolveEngine(s.Type)
	if err != nil {
		return err
	}

	err = e.RegisterSchema(registrant, s)
	return errors.Wrap(err, "error from credential engine")

}

func (r *Registry) CreateCredentialOffer(issuer *datastore.DID, s *datastore.Schema) (*decorator.AttachmentData, error) {
	e, err := r.resolveEngine(s.Type)
	if err != nil {
		return nil, err
	}

	return e.CreateCredentialOffer(issuer, s)
}

func (r *Registry) IssueCredential(issuer *datastore.DID, s *datastore.Schema) (*decorator.AttachmentData, error) {
	e, err := r.resolveEngine(s.Type)
	if err != nil {
		return nil, err
	}

	return e.IssueCredential(s, nil, nil, nil)
}

func (r *Registry) resolveEngine(method string) (CredentialEngine, error) {
	for _, e := range r.engines {
		if e.Accept(method) {
			return e, nil
		}
	}

	return nil, fmt.Errorf("credential type %s not supported for engine", method)
}

// WithEngine adds did method implementation for store.
func WithEngine(e CredentialEngine) Option {
	return func(opts *Registry) {
		opts.engines = append(opts.engines, e)
	}
}
