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

type value struct {
	attr     string
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

func (r *ValuesBuilder) AddKnown(attr string, decValue string) {
	r.known = append(r.known, value{attr, decValue})
}

func (r *ValuesBuilder) AddHidden(attr string, decValue string) {
	r.hidden = append(r.hidden, value{attr, decValue})
}

func (r *ValuesBuilder) AddCommitment(attr string, decValue string, blindingFactor string) {
	r.commitment = append(r.commitment, commit{attr, decValue, blindingFactor})
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
			return errors.Errorf("error from URSA adding hidden: %d", result)
		}
	}
	for _, h := range r.hidden {
		cattr := C.CString(h.attr)
		cval := C.CString(h.decValue)
		result = C.ursa_cl_credential_values_builder_add_dec_hidden(builder, cattr, cval)
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
			return errors.Errorf("error from URSA adding commitment %d", result)
		}
	}

	result = C.ursa_cl_credential_values_builder_finalize(builder, &r.values)
	return nil
}

func (r *ValuesBuilder) Free() error {
	result := C.ursa_cl_credential_values_free(r.values)
	if result != 0 {
		return errors.Errorf("error from URSA freeing values: %d", result)
	}
	return nil
}
