package indy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/schema"
	cursa "github.com/scoir/canis/pkg/ursa"
)

const (
	Format = "hlindy-zkp-v1.0"
)

type Engine struct {
	client VDRClient
	store  datastore.Store
	oracle Oracle
}

func New(prov Provider) (*Engine, error) {
	eng := &Engine{}

	var err error
	eng.client, err = prov.IndyVDR()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr for indy proof engine")
	}

	eng.store = prov.Store()
	eng.oracle = prov.Oracle()

	return eng, nil
}

func (r *Engine) Accept(format string) bool {
	return format == Format
}

// PresentationRequest to be encoded and sent as data in the RequestPresentation response
// Ref: https://github.com/hyperledger/indy-sdk/blob/57dcdae74164d1c7aa06f2cccecaae121cefac25/libindy/src/api/anoncreds.rs#L1214
type PresentationRequest struct {
	Name                string                                       `json:"name,omitempty"`
	Version             string                                       `json:"version,omitempty"`
	Nonce               string                                       `json:"nonce,omitempty"`
	RequestedAttributes map[string]*schema.IndyProofRequestAttr      `json:"requested_attributes,omitempty"`
	RequestedPredicates map[string]*schema.IndyProofRequestPredicate `json:"requested_predicates,omitempty"`
	NonRevoked          schema.NonRevokedInterval                    `json:"non_revoked,omitempty"`
}

// RequestPresentationAttach
func (r *Engine) RequestPresentation(name, version string, attrInfo map[string]*schema.IndyProofRequestAttr,
	predicateInfo map[string]*schema.IndyProofRequestPredicate) (*decorator.AttachmentData, error) {

	nonce, err := r.oracle.NewNonce()
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(&PresentationRequest{
		Name:                name,
		Version:             version,
		Nonce:               nonce,
		RequestedAttributes: attrInfo,
		RequestedPredicates: predicateInfo,
	})
	if err != nil {
		return nil, err
	}

	return &decorator.AttachmentData{
		Base64: base64.StdEncoding.EncodeToString(b),
	}, nil
}

// RequestPresentationFormat
func (r *Engine) RequestPresentationFormat() string {
	return Format
}

func (r *Engine) Verify(presentation, request []byte, theirDID string, myDID string) error {

	indyProof := &schema.IndyProof{}
	err := json.Unmarshal(presentation, indyProof)
	if err != nil {
		return errors.Wrap(err, "invalid presentation format, not indy proof")
	}

	proofRequest := &PresentationRequest{}
	err = json.Unmarshal(request, proofRequest)
	if err != nil {
		return errors.Wrap(err, "invalid proof request format")
	}

	receivedRevealedAttrs, err := receivedRevealedAttrs(indyProof)
	if err != nil {
		return err
	}

	receivedUnrevealedAttrs, err := receivedUnrevealedAttrs(indyProof)
	if err != nil {
		return err
	}

	receivedPredicates, err := receivedPredicates(indyProof)
	if err != nil {
		return err
	}

	receivedSelfAttestedAttrs := receivedSelfAttestedAttrs(indyProof)

	err = compareAttrFromProofAndRequest(proofRequest, receivedRevealedAttrs, receivedUnrevealedAttrs,
		receivedSelfAttestedAttrs, receivedPredicates)
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = verifyRevealedAttrubuteValues(proofRequest, indyProof)
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = verifiyRequesetedRestrictions(proofRequest, indyProof.RequestedProof, receivedRevealedAttrs, receivedUnrevealedAttrs,
		receivedPredicates, receivedSelfAttestedAttrs)
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = compareTimestampsFromProofAndRequest(proofRequest, receivedRevealedAttrs, receivedUnrevealedAttrs,
		receivedSelfAttestedAttrs, receivedPredicates)
	if err != nil {
		return errors.Wrap(err, "")
	}

	credDefs := map[string]*vdr.ClaimDefData{}
	for _, identifier := range indyProof.Identifiers {

		credDef, err := r.getCredDef(identifier.CredDefID)
		if err != nil {
			return errors.Wrapf(err, "unable to load cred def %s", identifier.CredDefID)
		}
		credDefs[identifier.CredDefID] = credDef

	}

	return r.verifyCryptoCredential(indyProof, proofRequest, credDefs)
}

func (r *Engine) getCredDef(credDefID string) (*vdr.ClaimDefData, error) {
	rply, err := r.client.GetCredDef(credDefID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cred def from ledger")
	}

	indyCredDef := &vdr.ClaimDefData{}
	d, _ := json.Marshal(rply.Data)
	err = json.Unmarshal(d, indyCredDef)
	if err != nil {
		return nil, errors.Wrap(err, "invalid reply from ledger for cred def")
	}

	return indyCredDef, nil
}

func (r *Engine) getAttrbutesForCredential(subProofIdx int, requestedProof *schema.IndyRequestedProof, proofRequest *PresentationRequest) []*schema.IndyProofRequestAttr {
	var revealedAttrs []*schema.IndyProofRequestAttr

	for attrReferent, rattr := range requestedProof.RevealedAttrs {
		pa, ok := proofRequest.RequestedAttributes[attrReferent]
		if subProofIdx == int(rattr.SubProofIndex) && ok {
			revealedAttrs = append(revealedAttrs, pa)
		}
	}

	for attrReferent, rgroup := range requestedProof.RevealedAttrGroups {
		pa, ok := proofRequest.RequestedAttributes[attrReferent]
		if subProofIdx == int(rgroup.SubProofIndex) && ok {
			revealedAttrs = append(revealedAttrs, pa)
		}
	}

	return revealedAttrs
}

func (r *Engine) getPredicatesForCredential(subProofIdx int, requestedProof *schema.IndyRequestedProof,
	proofRequest *PresentationRequest) []*schema.IndyProofRequestPredicate {

	var predicates []*schema.IndyProofRequestPredicate

	for predicateReferent, rpredicate := range requestedProof.Predicates {
		p, ok := proofRequest.RequestedPredicates[predicateReferent]
		if subProofIdx == int(rpredicate.SubProofIndex) && ok {
			predicates = append(predicates, p)
		}
	}

	return predicates
}

func (r *Engine) buildSubProofRequest(attrs []*schema.IndyProofRequestAttr,
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
		fmt.Println("adding revealed attr to sub proof", cursa.AttrCommonView(name))
		err := subProofBuilder.AddRevealedAttr(cursa.AttrCommonView(name))
		if err != nil {
			return nil, errors.Wrap(err, "unable to add revealed attribute")
		}
	}

	for _, predicate := range predicates {
		err = subProofBuilder.AddPredicate(cursa.AttrCommonView(predicate.Name), predicate.PType, predicate.PValue)
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
