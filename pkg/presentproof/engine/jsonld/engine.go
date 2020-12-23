package jsonld

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/presentproof"
	"github.com/scoir/canis/pkg/schema"
)

type provider interface {
	IndyVDR() (indy.IndyVDRClient, error)
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

// Accept type should be json-ld
func (r *Engine) Accept(typ string) bool {
	return typ == "json-ld"
}

// RequestPresentation
func (r *Engine) RequestPresentation(name, version string, attrInfo map[string]*schema.IndyProofRequestAttr,
	predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error) {
	log.Println(attrInfo)
	log.Println(predicateInfo)

	return nil, errors.New("not implemented")
}

func (r *Engine) RequestPresentationFormat() string {
	return "LDS/LD-Proof"
}

func (r *Engine) Verify(presentation, request []byte, theirDID string, myDID string) error {
	return errors.New("not implemented")
}
