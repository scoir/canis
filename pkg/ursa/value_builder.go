package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"unsafe"

	"github.com/pkg/errors"
)

type value struct {
	attr     string
	raw      interface{}
	decValue string
}

type commit struct {
	attr           string
	decValue       string
	blindingFactor string
}

type ValuesBuilder struct {
	hidden     []value
	known      []value
	commitment []commit
	values     unsafe.Pointer
}

func NewValuesBuilder() *ValuesBuilder {
	return &ValuesBuilder{}
}

func (r *ValuesBuilder) AddKnown(attr string, raw interface{}) {
	r.known = append(r.known, value{attr: attr, raw: raw, decValue: r.encodeValue(raw)})
}

func (r *ValuesBuilder) AddKnownDec(attr string, decValue string) {
	r.known = append(r.known, value{attr: attr, decValue: decValue})
}

func (r *ValuesBuilder) AddHidden(attr string, raw interface{}) {
	r.hidden = append(r.hidden, value{attr: attr, raw: raw, decValue: r.encodeValue(raw)})
}

func (r *ValuesBuilder) AddHiddenDec(attr string, decValue string) {
	r.hidden = append(r.hidden, value{attr: attr, decValue: decValue})
}

func (r *ValuesBuilder) AddCommitment(attr string, raw string, blindingFactor string) {
	r.commitment = append(r.commitment, commit{attr: attr, decValue: r.encodeValue(raw), blindingFactor: blindingFactor})
}

func (r *ValuesBuilder) AddCommitmentDec(attr string, decValue string, blindingFactor string) {
	r.commitment = append(r.commitment, commit{attr: attr, decValue: decValue, blindingFactor: blindingFactor})
}

func (r *ValuesBuilder) Values() unsafe.Pointer {
	return r.values
}

func (r *ValuesBuilder) Finalize() error {
	var builder unsafe.Pointer

	result := C.ursa_cl_credential_values_builder_new(&builder)
	if result != 0 {
		return errors.Errorf("error from URSA creating values builder: %d", result)
	}

	for _, h := range r.hidden {
		cattr := C.CString(h.attr)
		cval := C.CString(h.decValue)
		result = C.ursa_cl_credential_values_builder_add_dec_hidden(builder, cattr, cval)
		C.free(unsafe.Pointer(cattr))
		C.free(unsafe.Pointer(cval))
		if result != 0 {
			var errJson *C.char
			C.ursa_get_current_error(&errJson)
			defer C.free(unsafe.Pointer(errJson))
			return errors.Errorf("error from URSA adding hidden: %s", C.GoString(errJson))
		}
	}
	for _, h := range r.known {
		cattr := C.CString(h.attr)
		cval := C.CString(h.decValue)
		result = C.ursa_cl_credential_values_builder_add_dec_known(builder, cattr, cval)
		C.free(unsafe.Pointer(cattr))
		C.free(unsafe.Pointer(cval))
		if result != 0 {
			return errors.Errorf("error from URSA adding known: %d", result)
		}
	}
	for _, h := range r.commitment {
		cattr := C.CString(h.attr)
		cval := C.CString(h.decValue)
		cfac := C.CString(h.blindingFactor)
		result = C.ursa_cl_credential_values_builder_add_dec_commitment(builder, cattr, cval, cfac)
		C.free(unsafe.Pointer(cattr))
		C.free(unsafe.Pointer(cval))
		C.free(unsafe.Pointer(cfac))
		if result != 0 {
			return errors.Errorf("error from URSA building schema: %d", result)
		}
	}

	result = C.ursa_cl_credential_values_builder_finalize(builder, &r.values)
	if result != 0 {
		return errors.Errorf("error from URSA finalizing builder: %d", result)
	}

	return nil
}

func (r *ValuesBuilder) Free() error {
	result := C.ursa_cl_credential_values_free(r.values)
	if result != 0 {
		return errors.Errorf("error from URSA freeing values: %d", result)
	}
	return nil
}

func (r *ValuesBuilder) encodeValue(raw interface{}) string {
	var enc string

	switch v := raw.(type) {
	case nil:
		enc = r.toEncodedNumber("None")
	case string:
		i, err := strconv.Atoi(v)
		if err == nil && (i <= math.MaxInt32 && i >= math.MinInt32) {
			enc = v
		} else {
			enc = r.toEncodedNumber(v)
		}
	case bool:
		if v {
			enc = "1"
		} else {
			enc = "0"
		}
	case int32:
		enc = strconv.Itoa(int(v))
	case int64:
		if v <= math.MaxInt32 && v >= math.MinInt32 {
			enc = strconv.Itoa(int(v))
		} else {
			enc = r.toEncodedNumber(strconv.Itoa(int(v)))
		}
	case int:
		if v <= math.MaxInt32 && v >= math.MinInt32 {
			enc = strconv.Itoa(v)
		} else {
			enc = r.toEncodedNumber(strconv.Itoa(v))
		}
	case float64:
		if v == 0 {
			enc = r.toEncodedNumber("0.0")
		} else {
			enc = r.toEncodedNumber(fmt.Sprintf("%f", v))
		}
	default:
		//Not sure what to do with Go and unknown types...  this works for now
		enc = r.toEncodedNumber(fmt.Sprintf("%v", v))
	}

	return enc
}

func (r *ValuesBuilder) toEncodedNumber(raw string) string {
	b := []byte(raw)
	hasher := sha256.New()
	hasher.Write(b)

	sh := hasher.Sum(nil)
	i := new(big.Int)
	i.SetBytes(sh)

	return i.String()
}

type AttrValue struct {
	Raw     interface{} `json:"raw"`
	Encoded string      `json:"encoded"`
}

func (r *ValuesBuilder) MarshalJSON() ([]byte, error) {
	attrs := map[string]AttrValue{}

	for _, k := range r.known {
		attrs[k.attr] = AttrValue{Raw: k.raw, Encoded: k.decValue}
	}

	return json.Marshal(attrs)
}
