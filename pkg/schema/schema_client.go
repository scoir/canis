/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

import (
	"log"
	"strings"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/indy"
)

type Client struct {
	vdr indy.VDR
}

func New() *Client {
	return &Client{}
}

func (r *Client) CreateSchema(did, verkey, schemaName string, version string, schemaAttrs []string) (string, error) {
	schemaID, err := r.vdr.CreateSchema(did, verkey, schemaName, version, schemaAttrs)
	if err != nil {
		return "", errors.Wrap(err, "unable to create schema")
	}

	return schemaID, nil
}

func (r *Client) GetSchemaByID(schemaID string) (*indy.SchemaData, error) {
	var did, name, version string
	p := strings.Split(schemaID, ":")
	if len(p) != 3 {
		return nil, errors.New("schemaID is invalid")
	}

	did = p[0]
	name = p[1]
	version = p[2]

	return r.GetSchema(did, name, version)
}

func (r *Client) GetSchema(did, name, version string) (*indy.SchemaData, error) {
	schema, err := r.vdr.GetSchema(did, name, version)
	if err != nil {
		return nil, errors.Wrap(err, "error getting schema")
	}

	return schema, nil
}

func (r *Client) CreateCredentialDefinition(issuer, verkey, schemaID string) (string, error) {

	schema, err := r.GetSchemaByID(schemaID)
	if err != nil {
		return "", errors.Wrap(err, "unable to find schema to create cred def")
	}

	rr := map[string]interface{}{}
	primaryKey := map[string]interface{}{
		"r":     rr,
		"s":     "poop",
		"n":     "poop",
		"t":     "poop",
		"rctxt": "poop",
	}
	for _, name := range schema.AttrNames {
		rr[name] = "poop"
	}
	credDefId, err := r.vdr.CreateClaimDef(issuer, verkey, schema.SeqNo, primaryKey, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to create claim def")
	}

	return credDefId, nil
}

func (r *Client) GetCredentialDefinition(did, schemaID string) ([]string, error) {
	schema, err := r.GetSchemaByID(schemaID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find schema to find cred def")
	}
	log.Println("looking for schema", schema.SeqNo, "of ", schemaID, " in get cred def")
	credDef, err := r.vdr.GetClaimDef(did, schema.SeqNo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get claim def from vdr")
	}

	credDefResp := CredentialDefinitionIDsResult{
		CredentialDefinitionIDs: []string{credDef.ID},
	}

	return credDefResp.CredentialDefinitionIDs, nil
}
