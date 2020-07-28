package did

import "C"
import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

type KeyPair struct {
	vk, sk string
}

type MyDIDInfo struct {
	DID        string
	Seed       string
	Cid        bool
	MethodName string
}

func (r *KeyPair) Verkey() string {
	return r.vk
}

func (r *KeyPair) Priv() ed25519.PrivateKey {
	pk, _ := base58.Decode(r.sk)
	return pk
}

type DIDValue struct {
	DID    string
	Method string
}

func (r *DIDValue) String() string {
	if r.Method == "" {
		return fmt.Sprintf("did:%s", r.DID)
	}
	return fmt.Sprintf("did:%s:%s", r.Method, r.DID)
}

type DID struct {
	DIDVal DIDValue
	Verkey string
}

func (r *DID) String() string {
	return r.DIDVal.String()
}

func CreateMyDid(info *MyDIDInfo) (*DID, *KeyPair, error) {

	edseed, err := convertSeed(info.Seed)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get seed")
	}

	var pubkey ed25519.PublicKey
	var privkey ed25519.PrivateKey
	if len(edseed) == 0 {
		pubkey, privkey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, errors.Wrap(err, "error generating keypair")
		}
	} else {
		privkey = ed25519.NewKeyFromSeed(edseed)
		pubkey = privkey.Public().(ed25519.PublicKey)
	}

	var did string
	if info.DID != "" {
		did = info.DID
	} else if info.Cid {
		did = base58.Encode(pubkey[0:16])
	} else {
		did = base58.Encode(pubkey)
	}

	out := &DID{
		DIDVal: DIDValue{
			DID:    did,
			Method: info.MethodName,
		},
		Verkey: base58.Encode(pubkey),
	}

	return out, &KeyPair{vk: base58.Encode(pubkey), sk: base58.Encode(privkey)}, nil

}

func convertSeed(seed string) ([]byte, error) {
	if seed == "" {
		return []byte{}, nil
	}

	if len(seed) == ed25519.SeedSize {
		return []byte(seed), nil
	}

	if strings.HasSuffix(seed, "=") {
		var out = make([]byte, ed25519.SeedSize)
		c, err := base64.StdEncoding.Decode(out, []byte(seed))
		if err != nil || c != ed25519.SeedSize {
			return nil, errors.New("invalid base64 seed value")
		}
		return out, nil
	}

	if len(seed) == 2*ed25519.SeedSize {
		var out = make([]byte, ed25519.SeedSize)
		c, err := hex.Decode(out, []byte(seed))
		if err != nil || c != ed25519.SeedSize {
			return nil, errors.New("invalid hex seed value")
		}
		return out, nil
	}

	return []byte{}, nil
}
