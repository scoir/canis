package indy

import "C"
import (
	"encoding/base64"
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	ursaWrapper "github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	api "github.com/scoir/canis/pkg/didcomm/verifier/api/protogen"
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
	crypto cryptoProvider
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

	//todo: this needs to be better, crypto is unique to the engine, however this feels hacky
	eng.crypto = &ursaCrypto{}
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

// RequestPresentationAttach
func (r *Engine) RequestPresentationAttach(attrInfo map[string]*api.AttrInfo,
	predicateInfo map[string]*api.PredicateInfo) (string, error) {

	nonce, err := r.crypto.NewNonce()
	if err != nil {
		return "", err
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
		return "", err
	}


	return base64.StdEncoding.EncodeToString(b), nil
}

// RequestPresentationFormat
func (r *Engine) RequestPresentationFormat() string {
	return Format
}

type cryptoProvider interface {
	NewNonce() (string, error)
}

type ursaCrypto struct {
}

// NewNonce wraps ursa.NewNonce until we switch to the go wrapper
func (r *ursaCrypto) NewNonce() (string, error) {
	return ursaWrapper.NewNonce()
}
