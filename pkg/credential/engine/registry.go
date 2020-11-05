package engine

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

type CredentialEngine interface {
	Accept(format string) bool
	CreateSchema(issuer *datastore.DID, s *datastore.Schema) (string, error)
	RegisterSchema(registrant *datastore.DID, s *datastore.Schema) error
	CreateCredentialOffer(issuer *datastore.DID, subjectDID string, s *datastore.Schema, value []byte) (string, *decorator.AttachmentData, error)
	IssueCredential(issuerDID *datastore.DID, s *datastore.Schema, offerID string,
		requestAttachment decorator.AttachmentData, values map[string]interface{}) (*decorator.AttachmentData, error)
	GetSchemaForProposal(proposal []byte) (string, error)
}

//go:generate mockery -name=CredentialRegistry
type CredentialRegistry interface {
	CreateSchema(s *datastore.Schema) (string, error)
	RegisterSchema(registrant *datastore.DID, s *datastore.Schema) error
	CreateCredentialOffer(issuer *datastore.DID, subjectDID string, s *datastore.Schema, value []byte) (string, *decorator.AttachmentData, error)
	IssueCredential(issuer *datastore.DID, s *datastore.Schema, offerID string,
		requestAttachment decorator.AttachmentData, values map[string]interface{}) (*decorator.AttachmentData, error)
	GetSchemaForProposal(format string, data []byte) (string, error)
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
	e, err := r.resolveEngine(s.Format)
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
	e, err := r.resolveEngine(s.Format)
	if err != nil {
		return err
	}

	err = e.RegisterSchema(registrant, s)
	return errors.Wrap(err, "error from credential engine")

}

func (r *Registry) CreateCredentialOffer(issuer *datastore.DID, subjectDID string, s *datastore.Schema, value []byte) (string, *decorator.AttachmentData, error) {
	e, err := r.resolveEngine(s.Format)
	if err != nil {
		return "", nil, err
	}

	return e.CreateCredentialOffer(issuer, subjectDID, s, value)
}

func (r *Registry) IssueCredential(issuer *datastore.DID, s *datastore.Schema, offerID string, requestAttachment decorator.AttachmentData,
	values map[string]interface{}) (*decorator.AttachmentData, error) {
	e, err := r.resolveEngine(s.Format)
	if err != nil {
		return nil, err
	}

	return e.IssueCredential(issuer, s, offerID, requestAttachment, values)
}

func (r *Registry) GetSchemaForProposal(format string, data []byte) (string, error) {
	e, err := r.resolveEngine(format)
	if err != nil {
		return "", err
	}

	return e.GetSchemaForProposal(data)
}

func (r *Registry) resolveEngine(format string) (CredentialEngine, error) {
	for _, e := range r.engines {
		if e.Accept(format) {
			return e, nil
		}
	}

	return nil, fmt.Errorf("credential format %s not supported by any engine", format)
}

// WithEngine adds did method implementation for store.
func WithEngine(e CredentialEngine) Option {
	return func(opts *Registry) {
		opts.engines = append(opts.engines, e)
	}
}
