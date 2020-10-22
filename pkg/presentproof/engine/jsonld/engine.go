package jsonld

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/presentproof"
	api "github.com/scoir/canis/pkg/protogen/common"
	"github.com/scoir/canis/pkg/ursa"
)

type provider interface {
	IndyVDR() (indy.IndyVDRClient, error)
	KMS() kms.KeyManager
	StorageProvider() storage.Provider
	Issuer() ursa.Issuer
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
func (r *Engine) RequestPresentationAttach(attrInfo map[string]*api.AttrInfo, predicateInfo map[string]*api.PredicateInfo) (string, error) {
	log.Println(attrInfo)
	log.Println(predicateInfo)

	return "", nil
}

func (r *Engine) RequestPresentationFormat() string {
	return "LDS/LD-Proof"
}
