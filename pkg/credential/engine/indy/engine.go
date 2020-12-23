package indy

import "C"
import (
	"encoding/json"
	"fmt"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
	cursa "github.com/scoir/canis/pkg/ursa"
)

const Indy = "hlindy-zkp-v1.0"

type creddefWalletRecord struct {
	KeyCorrectnessProof map[string]interface{}
	PrivateKey          map[string]interface{}
}

type CredentialEngine struct {
	client VDRClient
	kms    kms.KeyManager
	store  Store
	oracle Oracle
}

func New(prov Provider) (*CredentialEngine, error) {
	eng := &CredentialEngine{}

	var err error
	eng.store, err = prov.StorageProvider().OpenStore("indy_engine")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open Store for indy engine")
	}

	eng.client, err = prov.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr for indy credential engine")
	}

	eng.kms = prov.KMS()
	eng.oracle = prov.Oracle()

	return eng, nil
}

func (r *CredentialEngine) Accept(format string) bool {
	return format == Indy
}

func (r *CredentialEngine) CreateSchema(issuer *datastore.DID, s *datastore.Schema) (string, error) {
	extId := fmt.Sprintf("%s:2:%s:%s", issuer.DID.MethodID(), s.Name, s.Version)
	rply, err := r.client.GetSchema(extId)
	if err == nil && rply.SeqNo > 0 {
		return extId, nil
	}

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
	reply, err := r.client.GetSchema(s.ExternalSchemaID)
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

	indycd := cursa.NewCredentailDefinition()

	names := make([]string, len(s.Attributes))
	for i, attr := range s.Attributes {
		names[i] = attr.Name
	}

	indycd.AddSchemaFields(names...)

	indycd.AddNonSchemaField("master_secret")

	err = indycd.Finalize()
	if err != nil {
		return errors.Wrap(err, "unable to finalize indy credential definition")
	}

	pubKeyDef, _ := indycd.PublicKey()
	pubKey, _ := pubKeyDef["p_key"].(map[string]interface{})

	credDefId, err := r.client.CreateClaimDef(registrant.DID.MethodID(), reply.SeqNo, pubKey, nil, mysig)
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
		return errors.Wrap(err, "error Store cred def private key and proof")
	}

	return nil
}

const (
	CLSignatureType = "CL"
	DefaultTag      = "default"
)

func (r *CredentialEngine) CreateCredentialOffer(issuer *datastore.DID, _ string, s *datastore.Schema, _ []byte) (string, *decorator.AttachmentData, error) {
	indySchema, err := r.client.GetSchema(s.ExternalSchemaID)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to find schema on ledger to create cred def")
	}

	credDefID := cursa.CredentialDefinitionID(issuer, indySchema.SeqNo, CLSignatureType, DefaultTag)

	rec, err := r.getCredDefRecord(credDefID)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to retrieve credential info from wallet")
	}

	offer, err := r.buildIndyOffer(s.ExternalSchemaID, credDefID, rec.KeyCorrectnessProof)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create ursa issuer")
	}

	offerID := uuid.New().URN()
	err = r.store.Put(offerID, []byte(offer.Base64))
	if err != nil {
		return "", nil, errors.Wrap(err, "unexpected error saving offer")
	}

	return offerID, offer, nil

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

	request := cursa.CredentialRequest{}
	d, err := requestAttachment.Fetch()
	if err != nil {
		return nil, errors.New("invalid attachment for issuing indy credential")
	}
	err = json.Unmarshal(d, &request)
	if err != nil {
		return nil, errors.Wrap(err, "invalid attachment JSON for issuing indy credential")
	}

	offer := &schema.IndyCredentialOffer{}
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

	return r.buildIndyCredential(
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

func (r *CredentialEngine) GetSchemaForProposal(proposal []byte) (string, error) {
	cp := &CredentialProposal{}
	err := json.Unmarshal(proposal, cp)
	if err != nil {
		return "", errors.Wrap(err, "invalid Indy credential proposal")
	}

	return cp.SchemaID, nil
}
