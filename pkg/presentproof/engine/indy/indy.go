package indy

import (
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/schema"
	cursa "github.com/scoir/canis/pkg/ursa"
)

func (r *Engine) verifyCryptoCredential(indyProof *schema.IndyProof, proofRequest *PresentationRequest, credDefs map[string]*vdr.ClaimDefData) error {

	nonCredSchema, err := cursa.BuildNonCredentialSchema()
	verifier, err := ursa.NewProofVerifier()
	if err != nil {
		return errors.Wrap(err, "")
	}

	for subProofIdx, identifier := range indyProof.Identifiers {

		sch, err := r.store.GetSchemaByExternalID(identifier.SchemaID)
		if err != nil {
			return errors.Wrapf(err, "unable to get schema for identifier %d", subProofIdx)
		}
		credDef := credDefs[identifier.CredDefID]

		attrsForCredential := r.getAttrbutesForCredential(subProofIdx, indyProof.RequestedProof, proofRequest)
		predicatesForCredential := r.getPredicatesForCredential(subProofIdx, indyProof.RequestedProof, proofRequest)

		credentialSchema, err := cursa.BuildCredentialSchema(sch.Attributes)
		if err != nil {
			return errors.Wrap(err, "unable to build verify schema")
		}

		subProofRequest, err := r.buildSubProofRequest(attrsForCredential, predicatesForCredential)

		pubKey, err := cursa.CredDefPublicKey(credDef.PKey(), credDef.RKey())
		if err != nil {
			return errors.Wrap(err, "unable to load cred def handle")
		}

		err = verifier.AddSubProofRequest(subProofRequest, credentialSchema, nonCredSchema, pubKey)
		if err != nil {
			return errors.Wrap(err, "")
		}

	}

	proofReqNonce, err := ursa.NonceFromJSON(proofRequest.Nonce)
	if err != nil {
		return errors.Wrap(err, "invalid proof request nonce")
	}

	cryptoProof, err := ursa.ProofFromJSON(indyProof.Proof)
	if err != nil {
		return errors.Wrap(err, "invalid ursa proof format")
	}
	defer func() { _ = cryptoProof.Free() }()

	return verifier.Verify(cryptoProof, proofReqNonce)
}
