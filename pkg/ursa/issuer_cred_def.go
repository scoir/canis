/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ursa

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"
)

const (
	DELIMITER = ":"
	MARKER    = "3"
)

func CredentialDefinitionID(did string, schemaID uint32, signatureType, tag string) string {
	return strings.Join([]string{did, MARKER, signatureType, strconv.Itoa(int(schemaID)), tag}, DELIMITER)
}

type CredentialDefinition struct {
	fields              []string
	nonfields           []string
	publicKey           string
	revocationKey       string
	privateKey          string
	keyCorrectnessProof string
}

func (r *CredentialDefinition) UrsaPublicKey() (*ursa.CredentialDefPubKey, error) {
	pubKey, err := ursa.CredentialPublicKeyFromJSON([]byte(r.publicKey))
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func (r *CredentialDefinition) KeyCorrectnessProof() (map[string]interface{}, error) {
	if r.keyCorrectnessProof == "" {
		return nil, errors.New("Finalize must be called after adding fields")
	}
	proofDef := map[string]interface{}{}
	err := json.Unmarshal([]byte(r.keyCorrectnessProof), &proofDef)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cl correctness proof")
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
	builder, err := ursa.NewCredentialSchemaBuilder()
	if err != nil {
		return errors.Errorf("error from URSA creating schema builder: %v", err)
	}

	for _, field := range r.fields {
		err := builder.AddAttr(AttrCommonView(field))
		if err != nil {
			return errors.Errorf("error adding field %s: %v", field, err)
		}
	}

	schema, err := builder.Finalize()
	if err != nil {
		return errors.Errorf("error from URSA building schema: %v", err)
	}
	defer schema.Free()

	nonBuilder, err := ursa.NewNonCredentialSchemaBuilder()
	if err != nil {
		return errors.Errorf("error from URSA creating non-schema: %v", err)
	}

	for _, field := range r.nonfields {
		err := nonBuilder.AddAttr(AttrCommonView(field))
		if err != nil {
			return errors.Errorf("error adding non-schema field: %v", err)
		}
	}

	nonSchema, err := nonBuilder.Finalize()
	if err != nil {
		return errors.Errorf("error from URSA finalizing non-schema: %v", err)
	}
	defer nonSchema.Free()

	credDef, err := ursa.NewCredentialDef(schema, nonSchema, false)
	if err != nil {
		return errors.Errorf("error creating new credential definition: %v", err)
	}

	pubJson, err := credDef.PubKey.ToJSON()
	if err != nil {
		return errors.Errorf("error from URSA turning pub key to json: %v", err)
	}
	defer credDef.PubKey.Free()

	privJson, err := credDef.PrivKey.ToJSON()
	if err != nil {
		return errors.Errorf("error from URSA turning private key to json: %v", err)
	}
	defer credDef.PrivKey.Free()

	proofJson, err := credDef.KeyCorrectnessProof.ToJSON()
	if err != nil {
		return errors.Errorf("error from URSA turning key correctness proof to json: %v", err)
	}
	defer credDef.KeyCorrectnessProof.Free()

	r.publicKey = string(pubJson)
	r.privateKey = string(privJson)
	r.keyCorrectnessProof = string(proofJson)

	return nil

}
