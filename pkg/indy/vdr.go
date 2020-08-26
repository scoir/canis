/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package indy

import (
	"encoding/json"
)

type VDR interface {
	CreateNym() (string, string, error)
	SetEndpoint(did, verkey, endpoint string) error
	CreateClaimDef(from, verkey string, ref uint32, primary, revocation map[string]interface{}) (string, error)
	CreateSchema(issuerDID, verkey, name, version string, attrs []string) (string, error)
	GetClaimDef(origin string, ref uint32) (*ClaimDefData, error)
	GetSchema(issuerDID, name, version string) (*SchemaData, error)
	GetSchemaByID(ID string) (*SchemaData, error)
}

type Operation struct {
	Type string `json:"type"`
}

type ClaimDefData struct {
	ID         string                 `json:"-"`
	Primary    map[string]interface{} `json:"primary"`
	Revocation map[string]interface{} `json:"revocation,omitempty"`
}

func (r *ClaimDefData) PKey() string {
	d, _ := json.MarshalIndent(r.Primary, " ", " ")
	return string(d)
}

func (r *ClaimDefData) RKey() string {
	d, _ := json.MarshalIndent(r.Revocation, " ", " ")
	return string(d)
}

type Schema struct {
	Operation `json:",inline"`
	Dest      string     `json:"dest"`
	Data      SchemaData `json:"data"`
}

type SchemaData struct {
	ID        string   `json:"id"`
	SeqNo     uint32   `json:"seq_no"`
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	AttrNames []string `json:"attr_names"`
}
