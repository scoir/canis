package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/base64"
	"encoding/json"
	"unsafe"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

type Credential struct {
	SchemaID                  string
	CredDefID                 string
	Values                    *ValuesBuilder
	Signature                 json.RawMessage
	SignatureCorrectnessProof json.RawMessage
}

type Issuer interface {
	IssueCredential(issuerDID string, schemaID, credDefID, offerNonce string, blindedMasterSecret, blindedMSCorrectnessProof, requestNonce string,
		credDef *vdr.ClaimDefData, credDefPrivateKey string, values map[string]interface{}) (*decorator.AttachmentData, error)
}

type IssuerServer struct {
}

func NewIssuer() *IssuerServer {
	return &IssuerServer{}
}

func (r *IssuerServer) IssueCredential(issuerDID string, schemaID, credDefID, offerNonce string, blindedMasterSecret, blindedMSCorrectnessProof, requestNonce string,
	credDef *vdr.ClaimDefData, credDefPrivateKey string, values map[string]interface{}) (*decorator.AttachmentData, error) {

	var blindedCredentialSecrets, blindedCredentialSecretsCorrectnessProof, credentialNonce,
		credentialIssuanceNonce, credentialValues, credentialPubKey, credentialPrivKey unsafe.Pointer
	var credSignature, credSignatureCorrectnessProof unsafe.Pointer

	did := C.CString(issuerDID)

	blindedCredentialSecrets, err := BlindedCredentialSecretsFromJSON(blindedMasterSecret)
	if err != nil {
		return nil, err
	}

	blindedCredentialSecretsCorrectnessProof, err = BlindedCredentialSecretsCorrectnessProofFromJSON(blindedMSCorrectnessProof)
	if err != nil {
		return nil, err
	}

	credentialIssuanceNonce, err = NonceFromJSON(offerNonce)
	if err != nil {
		return nil, err
	}

	credentialNonce, err = NonceFromJSON(requestNonce)
	if err != nil {
		return nil, err
	}

	builder := NewValuesBuilder()
	for k, v := range values {
		builder.AddKnown(k, v)
	}
	err = builder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create values")
	}
	credentialValues = builder.Values()
	defer builder.Free()

	credentialPubKey, err = CredDefHandle(credDef)
	if err != nil {
		return nil, err
	}

	credentialPrivKey, err = CredentialPrivateKeyFromJSON(credDefPrivateKey)
	if err != nil {
		return nil, err
	}

	result := C.ursa_cl_issuer_sign_credential(did, blindedCredentialSecrets, blindedCredentialSecretsCorrectnessProof, credentialIssuanceNonce,
		credentialNonce, credentialValues, credentialPubKey, credentialPrivKey, &credSignature, &credSignatureCorrectnessProof)
	if result != 0 {
		return nil, ursaError("signing credentials")
	}

	defer func() {
		C.free(unsafe.Pointer(did))
		C.free(blindedCredentialSecrets)
		C.free(blindedCredentialSecretsCorrectnessProof)
		C.free(credentialNonce)
		C.free(credentialIssuanceNonce)
		C.free(credentialPubKey)
		C.free(credentialPrivKey)
	}()

	var sigOut, proofOut *C.char
	result = C.ursa_cl_credential_signature_to_json(credSignature, &sigOut)
	defer C.free(unsafe.Pointer(sigOut))
	result = C.ursa_cl_signature_correctness_proof_to_json(credSignatureCorrectnessProof, &proofOut)
	defer C.free(unsafe.Pointer(proofOut))

	cred := &Credential{
		SchemaID:                  schemaID,
		CredDefID:                 credDefID,
		Values:                    builder,
		Signature:                 []byte(C.GoString(sigOut)),
		SignatureCorrectnessProof: []byte(C.GoString(proofOut)),
	}

	d, _ := json.Marshal(cred)
	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil
}
