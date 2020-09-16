package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"

type CredentialSigner struct {
}

func NewSigner(proverDID string) *CredentialSigner {
	return &CredentialSigner{}
}
