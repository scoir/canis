package ursa

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"
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

	signatureParams := &ursa.SignatureParams{
		ProverID: issuerDID,
		BlindedCredentialSecrets: blindedMasterSecret,
		BlindedCredentialSecretsCorrectnessProof: blindedMSCorrectnessProof,
		CredentialIssuanceNonce: offerNonce,
		CredentialNonce: requestNonce,
	}

	valueBuilder := NewValuesBuilder()
	builder, err  := ursa.NewValueBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "ursa error occurred creating builder")
	}

	for k, v := range values {
		err = ursa.AddDecKnown(builder, k, EncodeValue(v))
		if err != nil {
			return nil, errors.Wrap(err, "ursa error occurred adding known values to builder")
		}

		valueBuilder.AddKnown(k, v)
	}

	builderValues, err := ursa.FinalizeBuilder(builder)
	defer ursa.FreeCredentialValues(builderValues)
	if err != nil {
		return nil, errors.Wrap(err, "ursa error occurred finalizing builder")
	}

	signatureParams.CredentialValues = builderValues
	valueBuilder.values = builderValues


	signatureParams.CredentialPubKey = fmt.Sprintf(`{"p_key": %s, "r_key": %s}`, credDef.PKey(), credDef.RKey())
	signatureParams.CredentialPrivKey = credDefPrivateKey

	signature, proof, err := signatureParams.SignCredential()
	if err != nil {
		return nil, errors.Wrap(err, "ursa error occurred signing credential")
	}

	cred := &Credential{
		SchemaID:                  schemaID,
		CredDefID:                 credDefID,
		Values:                    valueBuilder,
		Signature:                 signature,
		SignatureCorrectnessProof: proof,
	}

	d, _ := json.Marshal(cred)
	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(d),
	}, nil
}
