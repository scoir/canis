package ursa

import (
	"fmt"

	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

func CredDefHandle(cd *vdr.ClaimDefData) (*ursa.CredentialDefPubKey, error) {
	j := fmt.Sprintf(`{"p_key": %s, "r_key": %s}`, cd.PKey(), cd.RKey())

	pubKey, err := ursa.CredentialPublicKeyFromJSON([]byte(j))
	if err != nil {
		return nil, errors.Wrap(err, "JSON marshal error for public key")
	}

	return pubKey, nil
}
