package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/pkg/errors"
)

func NewNonce() (string, error) {
	var nonce unsafe.Pointer

	result := C.ursa_cl_new_nonce(&nonce)
	if result != 0 {
		return "", errors.Errorf("new nonce returned error %d", result)
	}
	defer C.ursa_cl_nonce_free(nonce)

	var d *C.char
	result = C.ursa_cl_nonce_to_json(nonce, &d)
	if result != 0 {
		return "", errors.Errorf("eror turning nonce to json: %d", result)
	}
	str := C.GoString(d)
	defer C.free(unsafe.Pointer(d))

	fmt.Println(str)
	var out string
	err := json.Unmarshal([]byte(str), &out)
	if err != nil {
		return "", errors.Wrap(err, "unable to unmarshal nonce json")
	}

	return out, nil
}
