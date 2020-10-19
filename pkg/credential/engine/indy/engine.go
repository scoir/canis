package indy

import "C"
import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/ursa"
)

const Indy = "hlindy-zkp-v1.0"

type creddefWalletRecord struct {
	KeyCorrectnessProof map[string]interface{}
	PrivateKey          map[string]interface{}
}

type CredentialEngine struct {
	client VDRClient
	kms    kms.KeyManager
	store  store
	issuer UrsaIssuer
}

type provider interface {
	IndyVDR() (indy.IndyVDRClient, error)
	KMS() kms.KeyManager
	StorageProvider() storage.Provider
	Issuer() ursa.Issuer
}

type UrsaIssuer interface {
	IssueCredential(issuerDID string, schemaID, credDefID, offerNonce string, blindedMasterSecret, blindedMSCorrectnessProof, requestNonce string,
		credDef *vdr.ClaimDefData, credDefPrivateKey string, values map[string]interface{}) (*decorator.AttachmentData, error)
}

type store interface {
	Get(k string) ([]byte, error)
	Put(k string, v []byte) error
}

type VDRClient interface {
	CreateSchema(issuerDID, name, version string, attrs []string, signer vdr.Signer) (string, error)
	CreateClaimDef(from string, ref uint32, pubKey, revocation map[string]interface{}, signer vdr.Signer) (string, error)
	GetCredDef(credDefID string) (*vdr.ReadReply, error)
	GetSchema(schemaID string) (*vdr.ReadReply, error)
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

	eng.kms = prov.KMS()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load KMS in indy credential engine")
	}
	eng.issuer = prov.Issuer()

	return eng, nil
}

func (r *CredentialEngine) Accept(format string) bool {
	return format == Indy
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

func (r *CredentialEngine) CreateCredentialOffer(issuer *datastore.DID, subjectDID string, s *datastore.Schema, values map[string]interface{}) (string, *decorator.AttachmentData, error) {
	schema, err := r.client.GetSchema(s.ExternalSchemaID)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to find schema on ledger to create cred def")
	}

	credDefID := ursa.CredentialDefinitionID(issuer, schema.SeqNo, CLSignatureType, DefaultTag)

	rec, err := r.getCredDefRecord(credDefID)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to retrieve credential info from wallet")
	}

	nonce, err := ursa.NewNonce()
	if err != nil {
		return "", nil, errors.Wrap(err, "unexpected error creating nonce")
	}

	offer := ursa.CredentialOffer{
		SchemaID:            s.ExternalSchemaID,
		CredDefID:           credDefID,
		KeyCorrectnessProof: rec.KeyCorrectnessProof,
		Nonce:               nonce,
	}

	offerID := uuid.New().URN()
	d, err := json.Marshal(offer)
	if err != nil {
		return "", nil, errors.Wrap(err, "unexpect error marshalling offer into JSON")
	}
	err = r.store.Put(offerID, d)
	if err != nil {
		return "", nil, errors.Wrap(err, "unexpected error saving offer")
	}

	return offerID, &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil

}

func (r *CredentialEngine) getCredDefRecord(credDefID string) (*creddefWalletRecord, error) {
	d, err := r.store.Get(credDefID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cred def ID for this agent")
	}

	rec := &creddefWalletRecord{}
	_ = json.Unmarshal(d, rec)
	return rec, nil
}

func (r *CredentialEngine) IssueCredential(issuerDID *datastore.DID, s *datastore.Schema, offerID string,
	requestAttachment decorator.AttachmentData, values map[string]interface{}) (*decorator.AttachmentData, error) {

	request := ursa.CredentialRequest{}
	d, err := requestAttachment.Fetch()
	if err != nil {
		return nil, errors.New("invalid attachment for issuing indy credential")
	}
	err = json.Unmarshal(d, &request)
	if err != nil {
		return nil, errors.Wrap(err, "invalid attachment JSON for issuing indy credential")
	}

	offer := &ursa.CredentialOffer{}
	d, err = r.store.Get(offerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load existing indy offer")
	}
	err = json.Unmarshal(d, &offer)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error decoding indy stored offer")
	}

	rply, err := r.client.GetCredDef(offer.CredDefID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve cred def ledger record")
	}
	credDef := &vdr.ClaimDefData{}
	err = credDef.UnmarshalReadReply(rply)
	if err != nil {
		return nil, errors.Wrap(err, "invalid reply from indy retrieving cred def ledger record")
	}

	rec, err := r.getCredDefRecord(offer.CredDefID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve credential info from wallet")
	}

	credDefPrivateKey, _ := json.Marshal(rec.PrivateKey)

	return r.issuer.IssueCredential(
		issuerDID.DID.MethodID(),
		s.ExternalSchemaID,
		offer.CredDefID,
		offer.Nonce,
		request.BlindedMS,
		request.BlindedMSCorrectnessProof,
		request.Nonce,
		credDef,
		string(credDefPrivateKey),
		values,
	)

}
