package ursa

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	DELIMITER = ":"
	MARKER    = "3"
)

func CredentialDefinitionID(did *datastore.DID, schemaID uint32, signatureType, tag string) string {
	return strings.Join([]string{did.DID.MethodID(), MARKER, signatureType, strconv.Itoa(int(schemaID)), tag}, DELIMITER)
}

type CredentialDefinition struct {
	fields              []string
	nonfields           []string
	publicKey           string
	revocationKey       string
	privateKey          string
	keyCorrectnessProof string
}

func (r *CredentialDefinition) KeyCorrectnessProof() (map[string]interface{}, error) {
	if r.keyCorrectnessProof == "" {
		return nil, errors.New("Finalize must be called after adding fields")
	}
	proofDef := map[string]interface{}{}
	err := json.Unmarshal([]byte(r.keyCorrectnessProof), &proofDef)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cl pubkey")
	}

	return proofDef, nil
}

func (r *CredentialDefinition) PrivateKey() (map[string]interface{}, error) {
	if r.privateKey == "" {
		return nil, errors.New("Finalize must be called after adding fields")
	}
	privKeyDef := map[string]interface{}{}
	err := json.Unmarshal([]byte(r.privateKey), &privKeyDef)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cl pubkey")
	}

	return privKeyDef, nil
}

func (r *CredentialDefinition) PublicKey() (map[string]interface{}, error) {
	if r.publicKey == "" {
		return nil, errors.New("Finalize must be called after adding fields")
	}

	pubKeyDef := map[string]interface{}{}
	err := json.Unmarshal([]byte(r.publicKey), &pubKeyDef)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cl pubkey")
	}

	return pubKeyDef, nil
}

func NewCredentailDefinition() *CredentialDefinition {
	return &CredentialDefinition{}
}

func (r *CredentialDefinition) AddSchemaFields(f ...string) {
	r.fields = append(r.fields, f...)
}

func (r *CredentialDefinition) AddNonSchemaField(f ...string) {
	r.nonfields = append(r.nonfields, f...)
}

func (r *CredentialDefinition) Finalize() error {
	builder, err := ursa.CredentialSchemaBuilderNew()
	if err != nil {
		return errors.Wrap(err, "error from URSA creating schema builder")
	}

	for _, field := range r.fields {
		err = ursa.CredentialSchemaBuilderAddAttr(builder, field)
		if err != nil {
			return errors.Wrap(err, "error from URSA adding field")
		}
	}

	err = ursa.CredentialSchemaBuilderAddAttr(builder, "master_secret")
	if err != nil {
		return errors.Wrap(err, "error from URSA adding field")
	}

	schema, err := ursa.CredentialSchemaBuilderFinalize(builder)
	defer ursa.FreeCredentialSchema(builder)
	if err != nil {
		return errors.Wrap(err, "error from URSA finalizing builder")
	}

	nonBuilder, err := ursa.NonCredentialSchemaBuilderNew()
	if err != nil {
		return errors.Wrap(err, "error from URSA finalizing builder")
	}

	for _, field := range r.nonfields {
		err = ursa.NonCredentialSchemaBuilderAddAttr(nonBuilder, field)
		if err != nil {
			return errors.Wrap(err,"error adding non-schema field")
		}
	}

	nonSchema, err := ursa.NonCredentialSchemaBuilderFinalize(nonBuilder)
	defer ursa.FreeNonCredentialSchema(nonSchema)
	if err != nil {
		return errors.Wrap(err, "error from URSA finalizing nonbuilder")
	}

	credDef, err := ursa.NewCredentialDef(schema, nonSchema, false)
	if err != nil {
		return errors.Wrap(err, "error from URSA creating new cred def")
	}

	pubKey, err := ursa.CredentialPublicKeyToJSON(credDef.PubKey)
	defer ursa.FreeCredentialPublicKey(credDef.PubKey)
	if err != nil {
		return errors.Wrap(err, "error from URSA getting json pubkey")
	}

	privKey, err := ursa.CredentialPrivateKeyToJSON(credDef.PrivKey)
	defer ursa.FreeCredentialPrivateKey(credDef.PrivKey)
	if err != nil {
		return errors.Wrap(err, "error from URSA getting json privkey")
	}

	proof, err := ursa.CorrectnessProofToJSON(credDef.KeyCorrectnessProof)
	defer ursa.FreeCredentialKeyCorrectnessProof(credDef.KeyCorrectnessProof)
	if err != nil {
		return errors.Wrap(err, "error from URSA getting json correctness proof")
	}


	r.publicKey = string(pubKey)
	r.privateKey = string(privKey)
	r.keyCorrectnessProof = string(proof)

	return nil

}
