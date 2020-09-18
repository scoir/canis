package indy

import (
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

//go:generate mockery -name=IndyVDRClient
type IndyVDRClient interface {
	CreateClaimDef(from string, ref uint32, pubKey, revocation map[string]interface{}, signer vdr.Signer) (string, error)
	CreateNym(did, verkey, role, from string, signer vdr.Signer) error
	CreateAttrib(did, from string, data map[string]interface{}, signer vdr.Signer) error
	SetEndpoint(did, from string, ep string, signer vdr.Signer) error
	CreateSchema(issuerDID, name, version string, attrs []string, signer vdr.Signer) (string, error)
	Genesis() []byte
	Close() error
	Submit(request []byte) (*vdr.ReadReply, error)
	GetNym(did string) (*vdr.ReadReply, error)
	GetTxnAuthorAgreement() (*vdr.ReadReply, error)
	GetAcceptanceMethodList() (*vdr.ReadReply, error)
	GetEndpoint(did string) (*vdr.ReadReply, error)
	RefreshPool() error
	GetPoolStatus() (*vdr.PoolStatus, error)
	GetAttrib(did, raw string) (*vdr.ReadReply, error)
	GetSchema(schemaID string) (*vdr.ReadReply, error)
	GetCredDef(credDefID string) (*vdr.ReadReply, error)
	GetAuthRules() (*vdr.ReadReply, error)
	GetTxnTypeAuthRule(typ, action, field string) (*vdr.ReadReply, error)
	SubmitWrite(req *vdr.Request, signer vdr.Signer) (*vdr.WriteReply, error)
}
