/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

type SchemaIDsResult struct {
	SchemaIDs []string `json:"schema_ids"`
}

type CredentialDefinitionIDsResult struct {
	CredentialDefinitionIDs []string `json:"credential_definition_ids"`
}

type CredentialStatus struct {
	ID    string `bson:"ID" json:"id"`
	State string `bson:"State" json:"state"`
	Type  string `bson:"Type" json:"type"`
}

const (
	CredentialOffered  = "offered"
	CredentialIssued   = "issued"
	CredentialAccepted = "accepted"
)
