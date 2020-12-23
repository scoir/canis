package ursa

import (
	"fmt"
	"strings"

	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

type CryptoOracle struct{}

func (r *CryptoOracle) NewNonce() (string, error) {
	n, err := ursa.NewNonce()
	if err != nil {
		return "", err
	}

	js, err := n.ToJSON()
	return string(js), err
}

func CredDefPublicKey(pkey, rkey string) (*ursa.CredentialDefPubKey, error) {
	j := fmt.Sprintf(`{"p_key": %s, "r_key": %s}`, pkey, rkey)

	pubKey, err := ursa.CredentialPublicKeyFromJSON([]byte(j))
	if err != nil {
		return nil, errors.Wrap(err, "JSON marshal error for public key")
	}

	return pubKey, nil
}

func BuildNonCredentialSchema() (*ursa.NonCredentialSchemaHandle, error) {
	nonSchemaBuilder, err := ursa.NewNonCredentialSchemaBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create non cred schema builder")
	}

	err = nonSchemaBuilder.AddAttr("master_secret")
	if err != nil {
		return nil, errors.Wrap(err, "unable to add master secret")
	}

	return nonSchemaBuilder.Finalize()
}

func BuildCredentialSchema(attrs []*datastore.Attribute) (*ursa.CredentialSchemaHandle, error) {
	schemaBuilder, err := ursa.NewCredentialSchemaBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create schema builder")
	}

	for _, attr := range attrs {
		err := schemaBuilder.AddAttr(AttrCommonView(attr.Name))
		if err != nil {
			return nil, errors.Wrap(err, "unable to add schema attribute")
		}
	}

	return schemaBuilder.Finalize()
}

func AttrCommonView(attr string) string {
	return strings.ToLower(strings.Replace(attr, " ", "", -1))
}
