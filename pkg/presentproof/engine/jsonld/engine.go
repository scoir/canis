package jsonld

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	credindyengine "github.com/scoir/canis/pkg/credential/engine/indy"
	"github.com/scoir/canis/pkg/presentproof"
)

const (
	DIFPresentationExchange = "dif/presentation-exchange/definitions@v1.0"
	CanisOperationalDomain  = "canis.org/credentials"
)

type provider interface {
	IndyVDR() (credindyengine.VDRClient, error)
	KMS() kms.KeyManager
	StorageProvider() storage.Provider
}

type Engine struct {
	ctx      *ariescontext.Provider
	proofsup *presentproof.Supervisor
	subject  *didexchange.Connection
	store    store
}

type store interface {
	Get(k string) ([]byte, error)
	Put(k string, v []byte) error
}

func New(prov provider) (*Engine, error) {
	eng := &Engine{}

	var err error
	eng.store, err = prov.StorageProvider().OpenStore("jsonld_engine")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store for indy engine")
	}

	return eng, nil
}

// Accept type should be dif/presentation-exchange/definitions@v1.0
func (r *Engine) Accept(typ string) bool {
	return typ == DIFPresentationExchange
}

type RequestPresentation struct {
	// Domain is operational domain of a digital proof.
	Domain string `json:"domain,omitempty"`
	// Challenge is a random or pseudo-random value option authentication
	Challenge string `json:"challenge,omitempty"`

	Definitions *presexch.PresentationDefinitions `json:"presentation_definitions"`
}

// RequestPresentation
func (r *Engine) RequestPresentation(name string, definitions *presexch.PresentationDefinitions) (*decorator.AttachmentData, error) {

	rp := &RequestPresentation{
		Domain:      CanisOperationalDomain,
		Challenge:   uuid.New().String(),
		Definitions: definitions,
	}

	b, err := json.Marshal(rp)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error marshalling DIF presentation request")
	}

	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(b),
	}, nil

}

func (r *Engine) RequestPresentationFormat() string {
	return DIFPresentationExchange
}

func (r *Engine) Verify(presentation, request []byte, theirDID string, myDID string) error {
	return errors.New("not implemented")
}
