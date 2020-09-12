package mocks

import (
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
)

type MockKMS struct {
}

func (r *MockKMS) Create(kt kms.KeyType) (string, interface{}, error) {
	panic("implement me")
}

func (r *MockKMS) Get(keyID string) (interface{}, error) {
	return keyset.NewHandle(signature.ED25519KeyTemplate())
}

func (r *MockKMS) Rotate(kt kms.KeyType, keyID string) (string, interface{}, error) {
	panic("implement me")
}

func (r *MockKMS) ExportPubKeyBytes(keyID string) ([]byte, error) {
	panic("implement me")
}

func (r *MockKMS) CreateAndExportPubKeyBytes(kt kms.KeyType) (string, []byte, error) {
	return "id", []byte("abcdefghijklmnopqrstuvwxyz"), nil
}

func (r *MockKMS) PubKeyBytesToHandle(pubKey []byte, kt kms.KeyType) (interface{}, error) {
	panic("implement me")
}

func (r *MockKMS) ImportPrivateKey(privKey interface{}, kt kms.KeyType, opts ...kms.PrivateKeyOpts) (string, interface{}, error) {
	panic("implement me")
}
