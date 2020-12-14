/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

type ProofRequestWebRequest struct {
	ConnectionID string            `json:"connection_id"`
	ProofRequest *IndyProofRequest `json:"proof_request"`
}

type IndyProofRequest struct {
	Name                string                               `json:"name"`
	Version             string                               `json:"version"`
	Nonce               string                               `json:"nonce"`
	RequestedAttributes map[string]IndyProofRequestAttr      `json:"requested_attributes"`
	RequestedPredicates map[string]IndyProofRequestPredicate `json:"requested_predicates"`
}

type IndyProofRequestAttr struct {
	Name         string      `json:"name"`
	Names        []string    `json:"names"`
	Restrictions interface{} `json:"restrictions,omitempty"`
}

type IndyProofRequestPredicate struct {
	Name         string      `json:"name"`
	PType        string      `json:"p_type"`
	PValue       int32       `json:"p_value"`
	Restrictions interface{} `json:"restrictions"`
}
