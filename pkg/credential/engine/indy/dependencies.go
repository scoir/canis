package indy

import (
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

//go:generate mockery -inpkg -name=Provider
type Provider interface {
	IndyVDR() (VDRClient, error)
	KMS() kms.KeyManager
	StorageProvider() storage.Provider
	Oracle() Oracle
}

//go:generate mockery -name=Oracle
type Oracle interface {
	NewNonce() (string, error)
}

//go:generate mockery -name=VDRClient
type VDRClient interface {
	SetEndpoint(did, from string, ep string, signer vdr.Signer) error
	CreateSchema(issuerDID, name, version string, attrs []string, signer vdr.Signer) (string, error)
	CreateClaimDef(from string, ref uint32, pubKey, revocation map[string]interface{}, signer vdr.Signer) (string, error)
	GetCredDef(credDefID string) (*vdr.ReadReply, error)
	GetSchema(schemaID string) (*vdr.ReadReply, error)
	CreateNym(did, verkey, role, from string, signer vdr.Signer) error
	GetNym(did string) (*vdr.ReadReply, error)
}
