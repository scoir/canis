package indy

import (
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/schema"
	"github.com/scoir/canis/pkg/ursa"
)

func compareAttrFromProofAndRequest(proofReq *PresentationRequest, receivedRevealedAttrs map[string]*schema.Identifier,
	receivedUnrevealedAttrs map[string]*schema.Identifier, receivedSelfAttestedAttrs []string, receivedPredicates map[string]*schema.Identifier) error {

	empty := struct{}{}
	requestedAttrs := map[string]struct{}{}
	for k := range proofReq.RequestedAttributes {
		requestedAttrs[k] = empty
	}

	receivedAttrs := map[string]struct{}{}
	for k := range receivedRevealedAttrs {
		receivedAttrs[k] = empty
	}

	for k := range receivedUnrevealedAttrs {
		receivedAttrs[k] = empty
	}

	for _, k := range receivedSelfAttestedAttrs {
		receivedAttrs[k] = empty
	}

	if !reflect.DeepEqual(requestedAttrs, receivedAttrs) {
		return errors.Errorf("requested attributes [%v] do not correspond with received [%v]", requestedAttrs, receivedAttrs)
	}

	requestedPreds := map[string]struct{}{}
	for k := range proofReq.RequestedPredicates {
		requestedPreds[k] = empty
	}

	receivedPreds := map[string]struct{}{}
	for k := range receivedPredicates {
		receivedPreds[k] = empty
	}

	if !reflect.DeepEqual(requestedPreds, receivedPreds) {
		return errors.Errorf("requested predicates [%v] do not correspond to received [%v]", requestedPreds, receivedPreds)
	}

	return nil
}

func verifyRevealedAttrubuteValues(proofRequest *PresentationRequest, indyProof *schema.IndyProof) error {

	for attrReferent, info := range indyProof.RequestedProof.RevealedAttrs {
		requestAttr, ok := proofRequest.RequestedAttributes[attrReferent]
		if !ok {
			return errors.Errorf("attribute with referent %s not found in ProofRequests", attrReferent)
		}

		err := verifyRevealedAttrValue(requestAttr.Name, indyProof, info)
		if err != nil {
			return err
		}
	}

	for attrReferent, infos := range indyProof.RequestedProof.RevealedAttrGroups {
		requestAttr, ok := proofRequest.RequestedAttributes[attrReferent]
		if !ok {
			return errors.Errorf("attribute with referent %s not found in ProofRequests", attrReferent)
		}

		if len(infos.Values) != len(requestAttr.Names) {
			return errors.Errorf("Proof Revealed Attr Group does not match Proof Request Attribute Group, proof request attrs: %v, referent: %s, attr_infos: %v",
				proofRequest.RequestedAttributes, attrReferent, infos)
		}

		for _, attrName := range requestAttr.Names {
			attrInfo, ok := infos.Values[attrName]
			if !ok {
				return errors.Errorf("attribute with referent %s not found in ProofRequests", attrReferent)
			}

			err := verifyRevealedAttrValue(requestAttr.Name, indyProof, &schema.RevealedAttributeInfo{
				SubProofIndex: infos.SubProofIndex,
				Raw:           attrInfo.Raw,
				Encoded:       attrInfo.Encoded,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func verifyRevealedAttrValue(attrName string, proof *schema.IndyProof, info *schema.RevealedAttributeInfo) error {
	cryptoProof := &schema.CryptoProof{}
	err := json.Unmarshal(proof.Proof, cryptoProof)
	if err != nil {
		return errors.Wrap(err, "invalid crypto Proof")
	}

	revealedAttrEnc := info.Encoded
	subProofIdx := int(info.SubProofIndex)

	if subProofIdx >= len(cryptoProof.Proofs) {
		return errors.Errorf("Crypto not found by index %d", subProofIdx)
	}
	proofProof := cryptoProof.Proofs[subProofIdx]
	attrs := proofProof.RevealedAttrs()

	var cryptoProofEnc string
	for k, v := range attrs {
		if ursa.AttrCommonView(k) == ursa.AttrCommonView(attrName) {
			cryptoProofEnc = v
			break
		}
	}

	if cryptoProofEnc == "" {
		return errors.Errorf("Attribute with name \"%s\" not found in CryptoProof", attrName)
	}

	i := new(big.Int)
	j := new(big.Int)
	i.SetString(revealedAttrEnc, 10)
	j.SetString(cryptoProofEnc, 10)

	if i.Cmp(j) != 0 {
		return errors.Errorf("Encoded Values for \"%s\" are different in RequestedProof \"%s\" and CryptoProof \"%s\"", attrName, revealedAttrEnc, cryptoProofEnc)
	}

	return nil
}

func verifiyRequesetedRestrictions(proofReq *PresentationRequest, requestedProof *schema.IndyRequestedProof, receivedRevealedAttrs map[string]*schema.Identifier,
	receivedUnrevealedAttrs map[string]*schema.Identifier, receivedPredicates map[string]*schema.Identifier, receivedSelfAttestedAttrs []string) error {

	//TODO: implement restrictions
	//proofAttrIdentifiers := map[string]*schema.Identifier{}
	//for k, v := range receivedRevealedAttrs {
	//	proofAttrIdentifiers[k] = v
	//}
	//
	//for k, v := range receivedUnrevealedAttrs {
	//	proofAttrIdentifiers[k] = v
	//}
	//
	//requestedAttrs := map[string]*schema.IndyProofRequestAttr{}
	//for referent, info := range proofReq.RequestedAttributes {
	//	if !isSelfAttested(referent, info, receivedSelfAttestedAttrs) {
	//		requestedAttrs[referent] = info
	//	}
	//}
	//
	//for referent, info := range requestedAttrs {
	//
	//}

	return nil
}

func isSelfAttested(referent string, info *schema.IndyProofRequestAttr, attrs []string) bool {
	//TODO:  implement WQL checks for AND/OR
	for _, attr := range attrs {
		if attr == referent {
			return true
		}
	}
	return false
}

func compareTimestampsFromProofAndRequest(proofReq *PresentationRequest, receivedRevealedAttrs map[string]*schema.Identifier,
	receivedUnrevealedAttrs map[string]*schema.Identifier, receivedSelfAttestedAttrs []string, receivedPredicates map[string]*schema.Identifier) error {

	for referent, info := range proofReq.RequestedAttributes {
		err := validateTimestamp(receivedRevealedAttrs, referent, proofReq.NonRevoked, info.NonRevoked)
		if err != nil {
			err = validateTimestamp(receivedUnrevealedAttrs, referent, proofReq.NonRevoked, info.NonRevoked)
			if err != nil {
				found := false
				for _, attr := range receivedSelfAttestedAttrs {
					if attr == referent {
						found = true
						break
					}
				}

				if !found {
					return errors.New("invalid structure")
				}
			}
		}
	}

	for referent, predicate := range proofReq.RequestedPredicates {
		err := validateTimestamp(receivedPredicates, referent, proofReq.NonRevoked, predicate.NonRevoked)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateTimestamp(attrs map[string]*schema.Identifier, referent string, globalInterval schema.NonRevokedInterval,
	localInterval schema.NonRevokedInterval) error {

	if getNonRevocInterval(globalInterval, localInterval) == nil {
		return nil
	}

	_, ok := attrs[referent]
	if !ok {
		return errors.New("invalid structure")
	}

	return nil
}

func getNonRevocInterval(global schema.NonRevokedInterval, local schema.NonRevokedInterval) *schema.NonRevokedInterval {
	return &schema.NonRevokedInterval{}
}

func receivedRevealedAttrs(proof *schema.IndyProof) (map[string]*schema.Identifier, error) {
	out := map[string]*schema.Identifier{}
	for k, v := range proof.RequestedProof.RevealedAttrs {
		ident, err := getProofIdentifier(proof, int(v.SubProofIndex))
		if err != nil {
			return nil, err
		}
		out[k] = ident
	}

	for k, v := range proof.RequestedProof.RevealedAttrGroups {
		ident, err := getProofIdentifier(proof, int(v.SubProofIndex))
		if err != nil {
			return nil, err
		}
		out[k] = ident
	}

	return out, nil
}

func receivedUnrevealedAttrs(proof *schema.IndyProof) (map[string]*schema.Identifier, error) {
	out := map[string]*schema.Identifier{}
	for k, v := range proof.RequestedProof.UnrevealedAttrs {
		ident, err := getProofIdentifier(proof, int(v.SubProofIndex))
		if err != nil {
			return nil, err
		}
		out[k] = ident
	}

	return out, nil
}

func receivedPredicates(proof *schema.IndyProof) (map[string]*schema.Identifier, error) {
	out := map[string]*schema.Identifier{}
	for k, v := range proof.RequestedProof.Predicates {
		ident, err := getProofIdentifier(proof, int(v.SubProofIndex))
		if err != nil {
			return nil, err
		}
		out[k] = ident
	}

	return out, nil
}

func receivedSelfAttestedAttrs(proof *schema.IndyProof) []string {
	out := make([]string, len(proof.RequestedProof.SelfAttestedAttrs))
	i := 0
	for k := range proof.RequestedProof.SelfAttestedAttrs {
		out[i] = k
		i++
	}

	return out
}

func getProofIdentifier(proof *schema.IndyProof, idx int) (*schema.Identifier, error) {
	if idx >= len(proof.Identifiers) {
		return nil, errors.New("index out of range")
	}
	return proof.Identifiers[idx], nil
}
