package schema

import (
	"encoding/json"
)

type IndyProof struct {
	Proof          json.RawMessage     `json:"proof"`
	RequestedProof *IndyRequestedProof `json:"requested_proof"`
	Identifiers    []*Identifier       `json:"identifiers"`
}

type IndyRequestedProof struct {
	RevealedAttrs      map[string]*RevealedAttributeInfo      `json:"revealed_attrs"`
	RevealedAttrGroups map[string]*RevealedAttributeGroupInfo `json:"revealed_attr_groups"`
	SelfAttestedAttrs  map[string]string                      `json:"self_attested_attrs"`
	UnrevealedAttrs    map[string]*SubProofReferent           `json:"unrevealed_attrs"`
	Predicates         map[string]*SubProofReferent           `json:"predicates"`
}

type Identifier struct {
	SchemaID  string `json:"schema_id"`
	CredDefID string `json:"cred_def_id"`
	RevRegID  string `json:"rev_reg_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type SubProofReferent struct {
	SubProofIndex int32 `json:"sub_proof_index"`
}

type RevealedAttributeInfo struct {
	SubProofIndex int32  `json:"sub_proof_index"`
	Raw           string `json:"raw"`
	Encoded       string `json:"encoded"`
}

type RevealedAttributeGroupInfo struct {
	SubProofIndex int32 `json:"sub_proof_index"`
	Values        map[string]*IndyAttributeValue
}

type IndyAttributeValue struct {
	Raw     string `json:"raw"`
	Encoded string `json:"encoded"`
}

type IndyRequestedAttribute struct {
	CredID    string `json:"cred_id"`
	Timestamp int64  `json:"timestamp"`
	Revealed  bool   `json:"revealed"`
}

type ProvingCredentialKey struct {
	CredID    string `json:"cred_id"`
	Timestamp int64  `json:"timestamp"`
}

type IndyRequestedCredentials struct {
	SelfAttestedAttrs   map[string]string                  `json:"self_attested_attrs"`
	RequestedAttributes map[string]*IndyRequestedAttribute `json:"requested_attributes"`
	RequestedPredicates map[string]ProvingCredentialKey    `json:"requested_predicates"`
}

type IndyRequestedAttributeInfo struct {
	AttrReferent  string                `json:"attr_referent"`
	AttributeInfo *IndyProofRequestAttr `json:"attribute_info"`
	Revealed      bool                  `json:"revealed"`
}

type IndyRequestedPredicateInfo struct {
	PredicateReferent string                     `json:"predicate_referent"`
	PredicateInfo     *IndyProofRequestPredicate `json:"predicate_info"`
}
