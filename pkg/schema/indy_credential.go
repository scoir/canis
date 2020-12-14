package schema

import (
	"encoding/json"
)

type IndyCredential struct {
	SchemaID                  string               `json:"schema_id"`
	CredDefID                 string               `json:"cred_def_id"`
	RevRegID                  string               `json:"rev_reg_id"`
	Signature                 json.RawMessage      `json:"signature"`
	SignatureCorrectnessProof json.RawMessage      `json:"signature_correctness_proof"`
	Values                    IndyCredentialValues `json:"values"`
}

type IndyCredentialValues map[string]*IndyAttributeValue
