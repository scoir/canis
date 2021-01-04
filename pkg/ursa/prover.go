package ursa

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
)

type Prover struct {
	store storage.Store
}

type proverProvider interface {
	StorageProvider() storage.Provider
}

func NewProver(ctx proverProvider) (*Prover, error) {
	s, err := ctx.StorageProvider().OpenStore("wallet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store 'wallet'")
	}

	return &Prover{store: s}, nil
}

func (r *Prover) CreateMasterSecret(masterSecretID string) (string, error) {
	s, err := r.GetMasterSecret(masterSecretID)
	if err == nil {
		return s, nil
	}

	ms, err := ursa.NewMasterSecret()
	if err != nil {
		return "", errors.Wrap(err, "URSA error creating master secret")
	}
	defer ms.Free()

	js, err := ms.ToJSON()
	if err != nil {
		return "", errors.Wrap(err, "URSA error creating master secret JSON")
	}

	val := struct {
		MS string `json:"ms"`
	}{}

	_ = json.Unmarshal(js, &val)
	fmt.Println(val.MS)
	err = r.store.Put(masterSecretID, []byte(val.MS))
	if err != nil {
		return "", errors.Wrap(err, "unable to store new master secret")
	}

	return val.MS, nil
}

func (r *Prover) GetMasterSecret(masterSecretID string) (string, error) {
	ms, err := r.store.Get(masterSecretID)
	if err != nil {
		return "", errors.New("master secret not found")
	}

	return string(ms), nil
}

type CredentialRequest struct {
	ProverDID                 string `json:"prover_did"`
	CredDefID                 string `json:"cred_def_id"`
	BlindedMS                 string `json:"blinded_ms"`
	BlindedMSCorrectnessProof string `json:"blinded_ms_correctness_proof"`
	Nonce                     string `json:"nonce"`
}

type CredentialRequestMetadata struct {
	MasterSecretBlindingData string `json:"master_secret_blinding_data"`
	Nonce                    string `json:"nonce"`
	MasterSecretName         string `json:"master_secret_name"`
}

func (r *Prover) CreateCredentialRequest(proverDID string, credDef *vdr.ClaimDefData, offer *schema.IndyCredentialOffer, masterSecret string) (*CredentialRequest, *CredentialRequestMetadata, error) {
	credentialPubKey, err := CredDefPublicKey(credDef.PKey(), credDef.RKey())
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to build credential public key")
	}
	credValuesBuilder, err := ursa.NewValueBuilder()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create ursa values builder")
	}

	err = credValuesBuilder.AddDecHidden("master_secret", masterSecret)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to adding to ursa values builder")
	}

	credentialValues, err := credValuesBuilder.Finalize()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to finalize builder")
	}
	defer credentialValues.Free()

	kp := offer.KeyCorrectnessProof
	d, _ := json.Marshal(kp)
	keyCorrectnessProof, err := ursa.CredentialKeyCorrectnessProofFromJSON(d)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get correctness proof from JSON")
	}
	defer keyCorrectnessProof.Free()

	nonce, err := ursa.NonceFromJSON(fmt.Sprintf("\"%s\"", offer.Nonce))
	if err != nil {
		return nil, nil, err
	}
	defer nonce.Free()

	blindedSecrets, err := ursa.BlindCredentialSecrets(credentialPubKey, keyCorrectnessProof, nonce, credentialValues)
	if err != nil {
		return nil, nil, err
	}

	cr := &CredentialRequest{
		ProverDID: proverDID,
		CredDefID: credDef.ID,
	}

	js, err := blindedSecrets.Handle.ToJSON()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	cr.BlindedMS = string(js)
	defer blindedSecrets.Handle.Free()

	js, err = blindedSecrets.CorrectnessProof.ToJSON()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	defer blindedSecrets.CorrectnessProof.Free()
	cr.BlindedMSCorrectnessProof = string(js)

	reqNonce, err := ursa.NewNonce()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create nonce for cred request")
	}
	js, err = reqNonce.ToJSON()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	defer reqNonce.Free()
	cr.Nonce = string(js)

	md := &CredentialRequestMetadata{MasterSecretName: "master_secret", Nonce: cr.Nonce}

	js, err = blindedSecrets.BlindingFactor.ToJSON()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	defer blindedSecrets.BlindingFactor.Free()
	md.MasterSecretBlindingData = string(js)

	return cr, md, nil
}

func (r *Prover) CreateProof(credentials map[string]*schema.IndyCredential, proofReq *schema.IndyProofRequest, requestedCreds *schema.IndyRequestedCredentials,
	masterSecret string, schemas map[string]*datastore.Schema, credDefs map[string]*vdr.ClaimDefData /*revocation*/) (*schema.IndyProof, error) {

	pb, err := ursa.NewProofBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unexpected ursa error for new proof")
	}
	err = pb.AddCommonAttribute("master_secret")
	if err != nil {
		return nil, errors.Wrap(err, "unexpected ursa error adding master secret")
	}

	requestedProof := &schema.IndyRequestedProof{
		RevealedAttrs:      map[string]*schema.RevealedAttributeInfo{},
		RevealedAttrGroups: map[string]*schema.RevealedAttributeGroupInfo{},
		SelfAttestedAttrs:  requestedCreds.SelfAttestedAttrs,
		UnrevealedAttrs:    map[string]*schema.SubProofReferent{},
		Predicates:         map[string]*schema.SubProofReferent{},
	}

	credentialsForProving, err := prepareCredentialsForProving(requestedCreds, proofReq)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing credentials for new proof")
	}

	ursaNonSchema, err := buildNonCredentialSchema()
	if err != nil {
		return nil, errors.Wrap(err, "unable to build non credential schema")
	}

	subProofIdx := 0
	identifiers := make([]*schema.Identifier, len(credentialsForProving))
	for credKey, preparedCred := range credentialsForProving {
		credential, ok := credentials[credKey.CredID]
		if !ok {
			return nil, errors.Errorf("Credential not found by ID %s", credKey.CredID)
		}

		sschema, ok := schemas[credential.SchemaID]
		if !ok {
			return nil, errors.Errorf("Schema not found by ID %s", credential.SchemaID)
		}

		credDef, ok := credDefs[credential.CredDefID]
		if !ok {
			return nil, errors.Errorf("CredentialDefinition not found by ID %s", credential.CredDefID)
		}

		ursaSchema, err := buildSchema(sschema.Attributes)
		if err != nil {
			return nil, errors.Wrap(err, "building sschema")
		}

		ursaValues, err := buildCredentialValues(credential.Values, masterSecret)
		if err != nil {
			return nil, errors.Wrap(err, "building credential values")
		}

		ursaSubProof, err := buildSubproof(preparedCred.requestedAttrInfo, preparedCred.predicateInfo)
		if err != nil {
			return nil, errors.Wrap(err, "building subproof")
		}

		credDefPublicKey, err := CredDefPublicKey(credDef.PKey(), credDef.RKey())
		if err != nil {
			return nil, errors.Wrap(err, "getting public key")
		}

		ursaSignature, err := ursa.CredentialSignatureFromJSON(credential.Signature)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		err = pb.AddSubProofRequest(ursaSubProof, ursaSchema, ursaNonSchema, ursaSignature, ursaValues, credDefPublicKey)
		if err != nil {
			return nil, errors.Wrap(err, "error adding sub proof")
		}

		identifier := &schema.Identifier{
			SchemaID:  credential.SchemaID,
			CredDefID: credential.CredDefID,
			Timestamp: credKey.Timestamp,
		}

		identifiers[subProofIdx] = identifier

		err = updateRequestedProof(preparedCred.requestedAttrInfo, preparedCred.predicateInfo, proofReq, credential,
			int32(subProofIdx), requestedProof)
		if err != nil {
			return nil, err
		}

		subProofIdx++

	}

	ursaNone, err := ursa.NonceFromJSON(proofReq.Nonce)
	if err != nil {
		return nil, errors.Wrap(err, "error loading proof req nonce")
	}

	ursaProof, err := pb.Finalize(ursaNone)
	if err != nil {
		return nil, errors.Wrap(err, "error finalizing proof")
	}

	proofJSON, err := ursaProof.ToJSON()
	if err != nil {
		return nil, errors.Wrap(err, "error generating proof JSON")
	}
	return &schema.IndyProof{
		Proof:          proofJSON,
		RequestedProof: requestedProof,
		Identifiers:    identifiers,
	}, nil

}

func (r *Prover) ProcessCredentialSignature(cred *schema.IndyCredential, credRequest *CredentialRequest,
	masterSecret, masterSecretBlindingData, credDefPKey string) (string, error) {

	cryptoSignature, err := ursa.CredentialSignatureFromJSON(cred.Signature)
	if err != nil {
		return "", errors.Wrap(err, "unable to get credential signature from JSON")
	}

	cryptoSigKP, err := ursa.CredentialSignatureCorrectnessProofFromJSON(cred.SignatureCorrectnessProof)
	if err != nil {
		return "", errors.Wrap(err, "unable to get signature correctness proof from JSON")
	}

	issuanceNonce, err := ursa.NonceFromJSON(credRequest.Nonce)
	if err != nil {
		return "", errors.Wrap(err, "unable to get nonce from JSON")
	}

	pubKey, err := CredDefPublicKey(credDefPKey, "")
	if err != nil {
		return "", errors.Wrap(err, "unable to get cred def public key")
	}

	blindingFactor, err := ursa.CredentialSecretsBlindingFactorsFromJSON([]byte(masterSecretBlindingData))
	if err != nil {
		return "", errors.Wrap(err, "unable to get blinding factors")
	}

	valueBuilder, err := ursa.NewValueBuilder()
	if err != nil {
		return "", errors.Wrap(err, "unable to create values builder")
	}

	for name, value := range cred.Values {
		err = valueBuilder.AddDecKnown(name, value.Encoded)
		if err != nil {
			return "", errors.Wrap(err, "unable to add known dec")
		}
	}

	err = valueBuilder.AddDecHidden("master_secret", masterSecret)
	if err != nil {
		return "", errors.Wrap(err, "unable to add master secret")
	}

	values, err := valueBuilder.Finalize()
	if err != nil {
		return "", errors.Wrap(err, "unable to finalize values builder")
	}

	err = cryptoSignature.ProcessCredentialSignature(values, cryptoSigKP, blindingFactor, pubKey, issuanceNonce)
	if err != nil {
		return "", errors.Wrap(err, "unable to process signature")
	}

	sig, err := cryptoSignature.ToJSON()
	if err != nil {
		return "", errors.Wrap(err, "unable to get processed signature JSON")
	}

	return string(sig), nil
}

func updateRequestedProof(reqAttrs []schema.IndyRequestedAttributeInfo, predicates []schema.IndyRequestedPredicateInfo,
	proofReq *schema.IndyProofRequest, credential *schema.IndyCredential, subProofIdx int32, requestedProof *schema.IndyRequestedProof) error {

	for _, attr := range reqAttrs {
		if attr.Revealed {
			attribute := proofReq.RequestedAttributes[attr.AttrReferent]
			if len(attribute.Names) > 0 {
				values := map[string]*schema.IndyAttributeValue{}
				for _, name := range attribute.Names {
					attrValues, err := getCredentialValuesForAttribute(credential.Values, name)
					if err != nil {
						return err
					}
					values[name] = attrValues
				}

				requestedProof.RevealedAttrGroups[attr.AttrReferent] = &schema.RevealedAttributeGroupInfo{
					SubProofIndex: subProofIdx,
					Values:        values,
				}

			} else {
				attrValues, err := getCredentialValuesForAttribute(credential.Values, attribute.Name)
				if err != nil {
					return err
				}

				requestedProof.RevealedAttrs[attr.AttrReferent] = &schema.RevealedAttributeInfo{
					SubProofIndex: subProofIdx,
					Raw:           attrValues.Raw,
					Encoded:       attrValues.Encoded,
				}
			}
		} else {
			requestedProof.UnrevealedAttrs[attr.AttrReferent] = &schema.SubProofReferent{SubProofIndex: subProofIdx}
		}
	}

	for _, predicate := range predicates {
		requestedProof.Predicates[predicate.PredicateReferent] = &schema.SubProofReferent{SubProofIndex: subProofIdx}
	}

	return nil
}

func getCredentialValuesForAttribute(values schema.IndyCredentialValues, name string) (*schema.IndyAttributeValue, error) {
	for k, v := range values {
		if AttrCommonView(k) == AttrCommonView(name) {
			return v, nil
		}
	}
	return nil, errors.Errorf("Credential value not found for attribute %s", name)
}

func buildNonCredentialSchema() (*ursa.NonCredentialSchemaHandle, error) {
	nonSchemaBuilder, err := ursa.NewNonCredentialSchemaBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unable to build non credential schema")
	}

	err = nonSchemaBuilder.AddAttr("master_secret")
	if err != nil {
		return nil, errors.Wrap(err, "unable to add master secret attr")
	}

	nonSchema, err := nonSchemaBuilder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "unable to finalize non credential schema")
	}

	return nonSchema, nil

}

func buildSubproof(reqAttrs []schema.IndyRequestedAttributeInfo, reqPredicates []schema.IndyRequestedPredicateInfo) (*ursa.SubProofRequestHandle, error) {
	subProofBuilder, err := ursa.NewSubProofRequestBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create sub proof builder")
	}

	for _, attr := range reqAttrs {
		if attr.Revealed {
			if len(attr.AttributeInfo.Names) > 0 {
				for _, name := range attr.AttributeInfo.Names {
					err = subProofBuilder.AddRevealedAttr(AttrCommonView(name))
					if err != nil {
						return nil, errors.Wrap(err, "unable to add revealed attr")
					}
				}
			} else {
				err = subProofBuilder.AddRevealedAttr(AttrCommonView(attr.AttributeInfo.Name))
				if err != nil {
					return nil, errors.Wrap(err, "unable to add revealed attr")
				}
			}
		}
	}

	for _, predicate := range reqPredicates {
		err = subProofBuilder.AddPredicate(AttrCommonView(predicate.PredicateInfo.Name), predicate.PredicateInfo.PType, predicate.PredicateInfo.PValue)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add predicate")
		}
	}

	subProof, err := subProofBuilder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "error finalizing sub proof")
	}

	return subProof, nil
}

func buildCredentialValues(values schema.IndyCredentialValues, ms string) (*ursa.CredentialValues, error) {
	builder, err := ursa.NewValueBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "error creating ursa value builder")
	}
	for k, v := range values {
		err = builder.AddDecKnown(k, v.Encoded)
		if err != nil {
			return nil, errors.Wrap(err, "error adding ursa value")
		}
	}
	if len(ms) > 0 {
		fmt.Println("WE ARE ADDING MS", ms)
		err = builder.AddDecHidden("master_secret", ms)
		if err != nil {
			return nil, errors.Wrap(err, "error adding hidden master secret")
		}
	}

	value, err := builder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "error finalizing value builder")
	}

	return value, nil

}

func buildSchema(attributes []*datastore.Attribute) (*ursa.CredentialSchemaHandle, error) {
	schemaBuilder, err := ursa.NewCredentialSchemaBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "unable to creaet ursa schema builder")
	}

	for _, attr := range attributes {
		fmt.Println("adding attr", attr.Name, "to schema")
		err = schemaBuilder.AddAttr(attr.Name)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add ursa schema attribute")
		}
	}

	s, err := schemaBuilder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "unable to finalize ursa schema")
	}

	return s, err

}

type preparedCred struct {
	requestedAttrInfo []schema.IndyRequestedAttributeInfo
	predicateInfo     []schema.IndyRequestedPredicateInfo
}

func prepareCredentialsForProving(requestedCreds *schema.IndyRequestedCredentials, proofReq *schema.IndyProofRequest) (map[schema.ProvingCredentialKey]*preparedCred, error) {
	credentialsForProving := map[schema.ProvingCredentialKey]*preparedCred{}

	for attrReferent, requestedAttr := range requestedCreds.RequestedAttributes {
		attrInfo, ok := proofReq.RequestedAttributes[attrReferent]
		if !ok {
			return nil, errors.New("bad attribute")
		}

		reqAttrInfo := schema.IndyRequestedAttributeInfo{
			AttrReferent:  attrReferent,
			AttributeInfo: &attrInfo,
			Revealed:      requestedAttr.Revealed,
		}

		k := schema.ProvingCredentialKey{CredID: requestedAttr.CredID, Timestamp: requestedAttr.Timestamp}
		cp, ok := credentialsForProving[k]
		if !ok {
			cp = &preparedCred{
				requestedAttrInfo: []schema.IndyRequestedAttributeInfo{},
				predicateInfo:     []schema.IndyRequestedPredicateInfo{},
			}
			credentialsForProving[k] = cp
		}

		cp.requestedAttrInfo = append(cp.requestedAttrInfo, reqAttrInfo)
	}

	for predicateReferent, provingCredKey := range requestedCreds.RequestedPredicates {
		predicateInfo, ok := proofReq.RequestedPredicates[predicateReferent]
		if !ok {
			return nil, errors.New("bad attribute")
		}

		var reqPredicateInfo = schema.IndyRequestedPredicateInfo{
			PredicateReferent: predicateReferent,
			PredicateInfo:     &predicateInfo,
		}

		cp, ok := credentialsForProving[provingCredKey]
		if !ok {
			cp = &preparedCred{
				requestedAttrInfo: []schema.IndyRequestedAttributeInfo{},
				predicateInfo:     []schema.IndyRequestedPredicateInfo{},
			}
			credentialsForProving[provingCredKey] = cp
		}

		cp.predicateInfo = append(cp.predicateInfo, reqPredicateInfo)
	}

	return credentialsForProving, nil
}
