package lds

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	docutil "github.com/hyperledger/aries-framework-go/pkg/doc/util"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	LinkedDataSignature = "lds/ld-proof"
)

type store interface {
	Get(k string) ([]byte, error)
	Put(k string, v []byte) error
}

type CredentialEngine struct {
	kms     kms.KeyManager
	store   store
	vdriReg vdriapi.Registry
}

type provider interface {
	KMS() kms.KeyManager
	StorageProvider() storage.Provider
	VDRIRegistry() vdriapi.Registry
}

type credOffer struct {
	OfferID    string
	SubjectDID string
	Values     map[string]interface{}
}

func New(prov provider) (*CredentialEngine, error) {
	eng := &CredentialEngine{
		vdriReg: prov.VDRIRegistry(),
	}

	var err error
	eng.store, err = prov.StorageProvider().OpenStore("indy_engine")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store for indy engine")
	}
	eng.kms = prov.KMS()
	return eng, nil
}

func (r *CredentialEngine) Accept(format string) bool {
	return format == LinkedDataSignature
}

func (r *CredentialEngine) CreateSchema(_ *datastore.DID, _ *datastore.Schema) (string, error) {
	//NO-OP
	return "", nil
}

func (r *CredentialEngine) RegisterSchema(_ *datastore.DID, _ *datastore.Schema) error {
	// NO-OP
	return nil
}

func (r *CredentialEngine) CreateCredentialOffer(_ *datastore.DID, subjectDID string, s *datastore.Schema,
	values map[string]interface{}) (string, *decorator.AttachmentData, error) {

	out := values
	out["@type"] = []string{s.Type}
	out["@context"] = s.Context

	offerID := uuid.New().URN()
	offer := &credOffer{
		OfferID:    offerID,
		SubjectDID: subjectDID,
		Values:     out,
	}

	d, err := json.Marshal(offer)
	if err != nil {
		return "", nil, errors.Wrap(err, "unexpect error marshalling offer into JSON")
	}
	err = r.store.Put(offerID, d)
	if err != nil {
		return "", nil, errors.Wrap(err, "unexpected error saving offer")
	}

	d, _ = json.Marshal(values)
	return offerID, &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil

}

func (r *CredentialEngine) IssueCredential(issuerDID *datastore.DID, s *datastore.Schema, offerID string,
	request decorator.AttachmentData, values map[string]interface{}) (*decorator.AttachmentData, error) {

	offer := &credOffer{}
	d, err := r.store.Get(offerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load existing lds offer")
	}
	err = json.Unmarshal(d, &offer)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error decoding lds stored offer")
	}

	bits, _ := request.Fetch()
	reqVals := map[string]interface{}{}
	_ = json.Unmarshal(bits, &reqVals)

	if !reflect.DeepEqual(reqVals, offer.Values) {
		return nil, errors.New("requested values do not match original offer")
	}

	record := values
	record["@id"] = offer.SubjectDID

	vc := &verifiable.Credential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
		},
		ID: uuid.New().URN(),
		Types: []string{
			"VerifiableCredential",
		},
		Subject: record,
		Issuer: verifiable.Issuer{
			ID: issuerDID.DID.String(),
		},
		Issued:  docutil.NewTime(time.Now()),
		Schemas: []verifiable.TypedID{},
	}

	if s.Type != "" {
		vc.Types = append(vc.Types, s.Type)
	}
	vc.Context = append(vc.Context, s.Context...)

	err = r.signCred(vc, issuerDID, vc.Context)
	if err != nil {
		return nil, errors.Wrap(err, "error signing lds credentaisl")
	}

	d, _ = json.Marshal(vc)
	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil
}

//TODO:  figure out how to support alternate signature types
func (r *CredentialEngine) signCred(vc *verifiable.Credential, issuerDID *datastore.DID, context []string) error {

	doc, err := r.vdriReg.Resolve(issuerDID.DID.String())
	if err != nil {
		return errors.Wrap(err, "unable to resolve issuer did")
	}

	signer, err := r.newCryptoSigner(issuerDID.KeyPair.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get crypto signer for lds signature")
	}

	sigSuite := ed25519signature2018.New(
		suite.WithSigner(signer),
		suite.WithVerifier(ed25519signature2018.NewPublicKeyVerifier()))

	ldpContext := &verifiable.LinkedDataProofContext{
		SignatureType:           "Ed25519Signature2018",
		SignatureRepresentation: verifiable.SignatureProofValue,
		Suite:                   sigSuite,
		VerificationMethod:      fmt.Sprintf("%s#%s", issuerDID.DID.String(), doc.PublicKey[0].ID[1:]),
	}

	err = vc.AddLinkedDataProof(ldpContext)
	if err != nil {
		return errors.Wrap(err, "unable to add linked data proof")
	}

	return nil
}

func (r *CredentialEngine) newCryptoSigner(kid string) (*subtle.ED25519Signer, error) {
	priv, err := r.kms.Get(kid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find key set")
	}

	kh := priv.(*keyset.Handle)
	prim, err := kh.Primitives()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load signer primitives")
	}
	return prim.Primary.Primitive.(*subtle.ED25519Signer), nil

}
