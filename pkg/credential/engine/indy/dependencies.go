package indy

import (
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

//go:generate mockery -inpkg -name=Provider
type Provider interface {
	IndyVDR() (VDRClient, error)
	KMS() kms.KeyManager
	StorageProvider() StorageProvider
	Oracle() Oracle
}

//go:generate mockery -inpkg -name=StorageProvider
type StorageProvider interface {
	OpenStore(string) (Store, error)
}

//go:generate mockery -name=Oracle
type Oracle interface {
	NewNonce() (string, error)
}

//go:generate mockery -name=Store
type Store interface {
	Get(k string) ([]byte, error)
	Put(k string, v []byte) error
}

//go:generate mockery -name=VDRClient
type VDRClient interface {
	CreateSchema(issuerDID, name, version string, attrs []string, signer vdr.Signer) (string, error)
	CreateClaimDef(from string, ref uint32, pubKey, revocation map[string]interface{}, signer vdr.Signer) (string, error)
	GetCredDef(credDefID string) (*vdr.ReadReply, error)
	GetSchema(schemaID string) (*vdr.ReadReply, error)
}
