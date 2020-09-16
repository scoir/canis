package indy

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/ursa"
)

const Indy = "indy"

type creddefWalletRecord struct {
	KeyCorrectnessProof map[string]interface{}
	PrivateKey          map[string]interface{}
}

type CredentialEngine struct {
	client *vdr.Client
	kms    kms.KeyManager
	store  storage.Store
}

type provider interface {
	IndyVDR() (*vdr.Client, error)
	KMS() (kms.KeyManager, error)
	StorageProvider() storage.Provider
}

func New(prov provider) (*CredentialEngine, error) {
	eng := &CredentialEngine{}

	var err error
	eng.store, err = prov.StorageProvider().OpenStore("indy_engine")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store for indy engine")
	}

	eng.client, err = prov.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr for indy credential engine")
	}

	eng.kms, err = prov.KMS()

	return eng, nil
}

func (r *CredentialEngine) Accept(typ string) bool {
	return typ == Indy
}

func (r *CredentialEngine) CreateSchema(issuer *datastore.DID, s *datastore.Schema) (string, error) {
	attr := make([]string, len(s.Attributes))
	for i, a := range s.Attributes {
		attr[i] = a.Name
	}
	kh, err := r.kms.Get(issuer.KeyPair.ID)
	if err != nil {
		return "", errors.Wrap(err, "unable to get private key")
	}

	privKeyHandle := kh.(*keyset.Handle)
	prim, err := privKeyHandle.Primitives()
	if err != nil {
		return "", errors.Wrap(err, "unable to load signer primitives")
	}
	mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)

	ischema, err := r.client.CreateSchema(issuer.DID.MethodID(), s.Name, s.Version, attr, mysig)
	if err != nil {
		return "", errors.Wrap(err, "indy vdr client unable to create schema")
	}

	return ischema, nil
}

func (r *CredentialEngine) RegisterSchema(registrant *datastore.DID, s *datastore.Schema) error {
	schema, err := r.client.GetSchema(s.ExternalSchemaID)
	if err != nil {
		return errors.Wrap(err, "unable to find schema on ledger to create cred def")
	}
	kh, err := r.kms.Get(registrant.KeyPair.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get private key")
	}

	privKeyHandle := kh.(*keyset.Handle)
	prim, err := privKeyHandle.Primitives()
	if err != nil {
		return errors.Wrap(err, "unable to load signer primitives")
	}
	mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)

	indycd := ursa.NewCredentailDefinition()

	names := make([]string, len(s.Attributes))
	for i, attr := range s.Attributes {
		names[i] = attr.Name
	}
	indycd.AddSchemaFields(names...)
	err = indycd.Finalize()
	if err != nil {
		return errors.Wrap(err, "unable to finalize indy credential definition")
	}

	pubKeyDef, _ := indycd.PublicKey()
	pubKey, _ := pubKeyDef["p_key"].(map[string]interface{})

	credDefId, err := r.client.CreateClaimDef(registrant.DID.MethodID(), schema.SeqNo, pubKey, nil, mysig)
	if err != nil {
		return errors.Wrap(err, "unable to create claim def")
	}

	privKeyDef, _ := indycd.PrivateKey()
	keyProof, _ := indycd.KeyCorrectnessProof()
	rec := creddefWalletRecord{
		PrivateKey:          privKeyDef,
		KeyCorrectnessProof: keyProof,
	}

	d, _ := json.Marshal(rec)
	err = r.store.Put(credDefId, d)
	if err != nil {
		return errors.Wrap(err, "error store cred def private key and proof")
	}

	return nil
}

const (
	CLSignatureType = "CL"
	DefaultTag      = "default"
)

func (r *CredentialEngine) CreateCredentialOffer(issuer *datastore.DID, s *datastore.Schema) (*decorator.AttachmentData, error) {
	schema, err := r.client.GetSchema(s.ExternalSchemaID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find schema on ledger to create cred def")
	}

	credDefID := ursa.CredentialDefinitionID(issuer, schema.SeqNo, CLSignatureType, DefaultTag)

	d, err := r.store.Get(credDefID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cred def ID for this agent")
	}

	rec := &creddefWalletRecord{}
	_ = json.Unmarshal(d, rec)

	nonce, err := ursa.NewNonce()
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error creating nonce")
	}

	offer := ursa.CredentialOffer{
		SchemaID:            s.ExternalSchemaID,
		CredDefID:           credDefID,
		KeyCorrectnessProof: rec.KeyCorrectnessProof,
		Nonce:               nonce,
	}

	d, err = json.Marshal(offer)
	if err != nil {
		return nil, errors.Wrap(err, "unexpect error marshalling offer into JSON")
	}

	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil

}

func (r *CredentialEngine) IssueCredential(s *datastore.Schema, offer *ursa.CredentialOffer, request *ursa.CredentialRequest,
	values *ursa.CredentialValues) (*decorator.AttachmentData, error) {
	return nil, errors.New("not implemented")
}
