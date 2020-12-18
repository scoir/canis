package ursa

import (
	"fmt"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
)

type Verifier struct {
	store datastore.Store
}

func NewVerifier(store datastore.Store) *Verifier {
	return &Verifier{
		store: store,
	}
}

func (r *Verifier) VerifyCredential(indyProof *schema.IndyProof, requestedAttrs map[string]*schema.IndyProofRequestAttr,
	requestedPredicates map[string]*schema.IndyProofRequestPredicate, proofRequestNonce string, credDefs map[string]*vdr.ClaimDefData) error {

	nonCredSchema, err := BuildNonCredentialSchema()
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

		attrsForCredential := r.getAttrbutesForCredential(subProofIdx, indyProof.RequestedProof, requestedAttrs)
		predicatesForCredential := r.getPredicatesForCredential(subProofIdx, indyProof.RequestedProof, requestedPredicates)

		credentialSchema, err := BuildCredentialSchema(sch.Attributes)
		if err != nil {
			return errors.Wrap(err, "unable to build verify schema")
		}

		subProofRequest, err := r.buildSubProofRequest(attrsForCredential, predicatesForCredential)

		pubKey, err := CredDefPublicKey(credDef.PKey(), credDef.RKey())
		if err != nil {
			return errors.Wrap(err, "unable to load cred def handle")
		}

		err = verifier.AddSubProofRequest(subProofRequest, credentialSchema, nonCredSchema, pubKey)
		if err != nil {
			return errors.Wrap(err, "")
		}

	}

	proofReqNonce, err := ursa.NonceFromJSON(proofRequestNonce)

	cryptoProof, err := ursa.ProofFromJSON(indyProof.Proof)
	if err != nil {
		return errors.Wrap(err, "invalid ursa proof format")
	}
	defer func() { _ = cryptoProof.Free() }()

	return verifier.Verify(cryptoProof, proofReqNonce)
}

func (r *Verifier) getAttrbutesForCredential(subProofIdx int, requestedProof *schema.IndyRequestedProof,
	requestedAttrs map[string]*schema.IndyProofRequestAttr) []*schema.IndyProofRequestAttr {

	var revealedAttrs []*schema.IndyProofRequestAttr

	for attrReferent, rattr := range requestedProof.RevealedAttrs {
		pa, ok := requestedAttrs[attrReferent]
		if subProofIdx == int(rattr.SubProofIndex) && ok {
			revealedAttrs = append(revealedAttrs, pa)
		}
	}

	for attrReferent, rgroup := range requestedProof.RevealedAttrGroups {
		pa, ok := requestedAttrs[attrReferent]
		if subProofIdx == int(rgroup.SubProofIndex) && ok {
			revealedAttrs = append(revealedAttrs, pa)
		}
	}

	return revealedAttrs
}

func (r *Verifier) getPredicatesForCredential(subProofIdx int, requestedProof *schema.IndyRequestedProof,
	requestedPredicates map[string]*schema.IndyProofRequestPredicate) []*schema.IndyProofRequestPredicate {

	var predicates []*schema.IndyProofRequestPredicate

	for predicateReferent, rpredicate := range requestedProof.Predicates {
		p, ok := requestedPredicates[predicateReferent]
		if subProofIdx == int(rpredicate.SubProofIndex) && ok {
			predicates = append(predicates, p)
		}
	}

	return predicates
}

func (r *Verifier) buildSubProofRequest(attrs []*schema.IndyProofRequestAttr,
	predicates []*schema.IndyProofRequestPredicate) (*ursa.SubProofRequestHandle, error) {

	subProofBuilder, err := ursa.NewSubProofRequestBuilder()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var names []string
	for _, attr := range attrs {
		if len(attr.Name) > 0 {
			names = append(names, attr.Name)
		}

		for _, name := range attr.Names {
			names = append(names, name)
		}
	}

	for _, name := range names {
		fmt.Println("adding revealed attr to sub proof", AttrCommonView(name))
		err := subProofBuilder.AddRevealedAttr(AttrCommonView(name))
		if err != nil {
			return nil, errors.Wrap(err, "unable to add revealed attribute")
		}
	}

	for _, predicate := range predicates {
		err = subProofBuilder.AddPredicate(AttrCommonView(predicate.Name), predicate.PType, predicate.PValue)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add predicate to sub proof")
		}
	}

	subProofRequest, err := subProofBuilder.Finalize()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return subProofRequest, nil
}

func (r *Verifier) NewNonce() (string, error) {
	n, err := ursa.NewNonce()
	if err != nil {
		return "", err
	}

	js, err := n.ToJSON()
	return string(js), err

}
