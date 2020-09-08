package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"

	"github.com/pkg/errors"
)

type CredentialDefinition struct {
	fields              []string
	nonfields           []string
	publicKey           string
	revocationKey       string
	privateKey          string
	keyCorrectnessProof string
}

func (r *CredentialDefinition) KeyCorrectnessProof() (string, error) {
	if r.keyCorrectnessProof == "" {
		return "", errors.New("Finalize must be called after adding fields")
	}
	return r.keyCorrectnessProof, nil
}

func (r *CredentialDefinition) PrivateKey() (string, error) {
	if r.privateKey == "" {
		return "", errors.New("Finalize must be called after adding fields")
	}
	return r.privateKey, nil
}

func (r *CredentialDefinition) PublicKey() (string, error) {
	if r.publicKey == "" {
		return "", errors.New("Finalize must be called after adding fields")
	}
	return r.publicKey, nil
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
	var builder, schema, nonbuilder, nonschema unsafe.Pointer
	result := C.ursa_cl_credential_schema_builder_new(&builder)
	if result != 0 {
		return errors.Errorf("error from URSA creating schema builder: %d", result)
	}

	for _, field := range r.fields {
		cfield := C.CString(field)
		result = C.ursa_cl_credential_schema_builder_add_attr(builder, cfield)
		C.free(unsafe.Pointer(cfield))
		if result != 0 {
			return errors.Errorf("error adding field %s: %d", field, result)
		}
	}

	result = C.ursa_cl_credential_schema_builder_finalize(builder, &schema)
	if result != 0 {
		return errors.Errorf("error from URSA building schema: %d", result)
	}

	result = C.ursa_cl_non_credential_schema_builder_new(&nonbuilder)
	if result != 0 {
		return errors.Errorf("error from URSA creating non-schema: %d", result)
	}
	for _, field := range r.nonfields {

		cfield := C.CString(field)
		result = C.ursa_cl_non_credential_schema_builder_add_attr(nonbuilder, cfield)
		C.free(unsafe.Pointer(cfield))
		if result != 0 {
			return errors.Errorf("error adding non-schema field: %d", result)
		}
	}

	result = C.ursa_cl_non_credential_schema_builder_finalize(nonbuilder, &nonschema)
	if result != 0 {
		return errors.Errorf("error from URSA finalizing non-schema: %d", result)
	}

	var credpub, credpriv, credproof unsafe.Pointer

	credresult := C.ursa_cl_issuer_new_credential_def(schema, nonschema, false, &credpub, &credpriv, &credproof)
	if credresult != 0 {
		return errors.Errorf("error from URSA creating cred def: %d", credresult)
	}

	var proofJson, pubJson, privJson *C.char
	credresult = C.ursa_cl_credential_public_key_to_json(credpub, &pubJson)
	if credresult != 0 {
		return errors.Errorf("error from URSA turning pub key to json: %d", credresult)
	}

	credresult = C.ursa_cl_credential_private_key_to_json(credpriv, &privJson)
	if credresult != 0 {
		return errors.Errorf("error from URSA turning private key to json: %d", credresult)
	}

	credresult = C.ursa_cl_credential_key_correctness_proof_to_json(credproof, &proofJson)
	if credresult != 0 {
		return errors.Errorf("error from URSA turning key correctness proof to json: %d", credresult)
	}

	C.ursa_cl_credential_schema_free(schema)
	C.ursa_cl_non_credential_schema_free(nonschema)
	C.ursa_cl_credential_private_key_free(credpriv)
	C.ursa_cl_credential_public_key_free(credpub)
	C.ursa_cl_credential_key_correctness_proof_free(credproof)

	r.publicKey = C.GoString(pubJson)
	r.privateKey = C.GoString(privJson)
	r.keyCorrectnessProof = C.GoString(proofJson)

	return nil

}
