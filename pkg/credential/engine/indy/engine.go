package indy

import (
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/crypto"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

const Indy = "indy"

type CredentialEngine struct {
	client *vdr.Client
	kms    kms.KeyManager
}

type provider interface {
	IndyVDR() (*vdr.Client, error)
	KMS() (kms.KeyManager, error)
}

func New(prov provider) (*CredentialEngine, error) {
	eng := &CredentialEngine{}

	var err error
	eng.client, err = prov.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr for indy credential engine")
	}

	eng.kms, err = prov.KMS()

	return eng, nil
}

func (r *CredentialEngine) Accept(typ string) bool {
	return typ == Indy
}

func (r *CredentialEngine) CreateSchema(issuer *datastore.DID, s *datastore.Schema) (string, error) {
	attr := make([]string, len(s.Attributes))
	for i, a := range s.Attributes {
		attr[i] = a.Name
	}
	mysig := crypto.NewSigner(issuer.KeyPair.RawPublicKey(), issuer.KeyPair.RawPrivateKey())

	ischema, err := r.client.CreateSchema(issuer.DID.MethodID(), s.Name, s.Version, attr, mysig)
	if err != nil {
		return "", errors.Wrap(err, "indy vdr client unable to create schema")
	}

	return ischema, nil
}

func (r *CredentialEngine) RegisterSchema(issuer *datastore.DID, registrant *datastore.DID, s *datastore.Schema) error {
	panic("implement me")
}

func (r *CredentialEngine) IssueCredential(s *datastore.Schema, c *api.Credential) (string, error) {
	panic("implement me")
}
