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

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"
)

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
