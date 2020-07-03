/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

import (
	"net/url"
)

type ByCredDef func(q url.Values)

func ByCredDefSchemaName(name string) ByCredDef {
	return func(q url.Values) {
		q.Set("schema_name", name)
	}
}
func ByCredDefSchemaID(ID string) ByCredDef {
	return func(q url.Values) {
		q.Set("schema_id", ID)
	}
}
func ByCredDefSchemaIssuerDID(did string) ByCredDef {
	return func(q url.Values) {
		q.Set("schema_issuer_did", did)
	}
}
func ByCredDefVersion(version string) ByCredDef {
	return func(q url.Values) {
		q.Set("schema_version", version)
	}
}
func ByCredDefIssuerDID(did string) ByCredDef {
	return func(q url.Values) {
		q.Set("issuer_did", did)
	}
}
func ByCredDefID(id string) ByCredDef {
	return func(q url.Values) {
		q.Set("cred_def_id", id)
	}
}

type BySchema func(q url.Values)

func BySchemaName(name string) BySchema {
	return func(q url.Values) {
		q.Set("schema_name", name)
	}
}
func BySchemaID(ID string) BySchema {
	return func(q url.Values) {
		q.Set("schema_id", ID)
	}
}
func BySchemaIssuerDID(did string) BySchema {
	return func(q url.Values) {
		q.Set("schema_issuer_did", did)
	}
}
func BySchemaVersion(version string) BySchema {
	return func(q url.Values) {
		q.Set("schema_version", version)
	}
}
