package indy

import "C"
import (
	"encoding/json"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"github.com/scoir/canis/pkg/didcomm/verifier/api"
	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/ursa"
)

const (
	Indy   = "indy"
	Format = "hlindy-zkp-v1.0"
)

type Engine struct {
	client VDRClient
	kms    kms.KeyManager
	store  store
}

type provider interface {
	IndyVDR() (indy.IndyVDRClient, error)
	KMS() kms.KeyManager
	StorageProvider() storage.Provider
	Verifier() ursa.Verifier
}

type store interface {
	Get(k string) ([]byte, error)
	Put(k string, v []byte) error
}

type VDRClient interface {
}

func New(prov provider) (*Engine, error) {
	eng := &Engine{}

	var err error
	eng.store, err = prov.StorageProvider().OpenStore("indy_engine")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store for indy engine")
	}

	eng.client, err = prov.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr for indy proof engine")
	}

	eng.kms = prov.KMS()
	return eng, nil
}

func (r *Engine) Accept(typ string) bool {
	return typ == Indy
}

// PresentationRequest to be encoded and sent as data in the RequestPresentation response
// Ref: https://github.com/hyperledger/indy-sdk/blob/57dcdae74164d1c7aa06f2cccecaae121cefac25/libindy/src/api/anoncreds.rs#L1214
type PresentationRequest struct {
	Name                string                        `json:"name,omitempty"`
	Version             string                        `json:"version,omitempty"`
	Nonce               string                        `json:"nonce,omitempty"`
	RequestedAttributes map[string]*api.AttrInfo      `json:"requested_attributes,omitempty"`
	RequestedPredicates map[string]*api.PredicateInfo `json:"requested_predicates,omitempty"`
	NonRevoked          string                        `json:"non_revoked,omitempty"`
}

func (r *Engine) RequestPresentationAttach(attrInfo map[string]*api.AttrInfo,
	predicateInfo map[string]*api.PredicateInfo) ([]byte, error) {

	nonce, err := ursa.NewNonce()
	if err != nil {
		return nil, err
	}

	//TODO: proper names and version
	b, err := json.Marshal(&PresentationRequest{
		Name:                "Proof name...",
		Version:             "0.0.1",
		Nonce:               nonce,
		RequestedAttributes: attrInfo,
		RequestedPredicates: predicateInfo,
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *Engine) RequestPresentationFormat() string {
	return Format
}
