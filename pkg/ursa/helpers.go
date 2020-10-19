package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

func NonceFromJSON(jsn string) (unsafe.Pointer, error) {
	var handle unsafe.Pointer
	cjson := C.CString(fmt.Sprintf(`"%s"`, jsn))
	defer C.free(unsafe.Pointer(cjson))

	result := C.ursa_cl_nonce_from_json(cjson, &handle)
	if result != 0 {
		return nil, ursaError("nonce")
	}

	return handle, nil
}

func CredentialKeyCorrectnessProofFromJSON(jsn string) (unsafe.Pointer, error) {
	var handle unsafe.Pointer
	cjson := C.CString(jsn)
	defer C.free(unsafe.Pointer(cjson))

	result := C.ursa_cl_credential_key_correctness_proof_from_json(cjson, &handle)
	if result != 0 {
		return nil, ursaError("credential key correctness proof")
	}

	return handle, nil
}

func BlindedCredentialSecretsFromJSON(jsn string) (unsafe.Pointer, error) {
	var handle unsafe.Pointer
	cjson := C.CString(jsn)
	defer C.free(unsafe.Pointer(cjson))

	result := C.ursa_cl_blinded_credential_secrets_from_json(cjson, &handle)
	if result != 0 {
		return nil, ursaError("blinded credential secrets")
	}

	return handle, nil
}

func BlindedCredentialSecretsCorrectnessProofFromJSON(jsn string) (unsafe.Pointer, error) {
	var handle unsafe.Pointer
	cjson := C.CString(jsn)
	defer C.free(unsafe.Pointer(cjson))

	result := C.ursa_cl_blinded_credential_secrets_correctness_proof_from_json(cjson, &handle)
	if result != 0 {
		return nil, ursaError("blinded credential secrets correctness proof")
	}

	return handle, nil
}

func CredentialPrivateKeyFromJSON(jsn string) (unsafe.Pointer, error) {
	var handle unsafe.Pointer
	cjson := C.CString(jsn)
	defer C.free(unsafe.Pointer(cjson))

	result := C.ursa_cl_credential_private_key_from_json(cjson, &handle)
	if result != 0 {
		return nil, ursaError("credential private key")
	}

	return handle, nil
}

func CredentialPublicKeyFromJSON(jsn string) (unsafe.Pointer, error) {
	var handle unsafe.Pointer
	cjson := C.CString(jsn)
	defer C.free(unsafe.Pointer(cjson))

	result := C.ursa_cl_credential_public_key_from_json(cjson, &handle)
	if result != 0 {
		return nil, ursaError("credential public key")
	}

	return handle, nil
}

func ursaError(msg string) error {
	var errJson *C.char
	C.ursa_get_current_error(&errJson)
	defer C.free(unsafe.Pointer(errJson))
	return errors.Errorf("error from URSA %s: %s", msg, C.GoString(errJson))
}

func CredDefHandle(cd *vdr.ClaimDefData) (unsafe.Pointer, error) {
	var out unsafe.Pointer
	j := fmt.Sprintf(`{"p_key": %s, "r_key": %s}`, cd.PKey(), cd.RKey())

	cj := C.CString(j)
	defer C.free(unsafe.Pointer(cj))
	result := C.ursa_cl_credential_public_key_from_json(cj, &out)
	if result != 0 {
		var errJson *C.char
		C.ursa_get_current_error(&errJson)
		defer C.free(unsafe.Pointer(errJson))
		return nil, errors.Errorf("error from URSA generating cred public key: %s", C.GoString(errJson))
	}

	return out, nil
}
