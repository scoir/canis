/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

import "github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

const (
	ScoirHighSchoolSchemaName = "Scoir High School"
	ScoirHighSchoolSchemaVer  = "0.0.1"
	ScoirCollegeSchemaName    = "Scoir College"
	ScoirCollegeSchemaVer     = "0.0.1"
)

var (
	ScoirHighSchoolSchemaAttrs = []string{
		"Name",
		"City",
		"State",
		"ZipCode",
		"CEEB",
	}

	ScoirCollegeSchemaAttrs = []string{
		"Name",
		"City",
		"State",
		"ZipCode",
		"SCID",
		"SATNumber",
		"ACTNumber",
	}
)

type ScoirHighSchoolCredential struct {
	*verifiable.Credential
}

type HighSchoolSubject struct {
	ID      string `json:"id"`
	Name    string `json:"Name"`
	City    string `json:"City"`
	State   string `json:"State"`
	ZipCode string `json:"ZipCode"`
	CEEB    string `json:"CEEB"`
}

type ScoirCollegeCredential struct {
	*verifiable.Credential
}

type CollegeSubject struct {
	ID        string `json:"id"`
	Name      string `json:"Name"`
	City      string `json:"City"`
	State     string `json:"State"`
	ZipCode   string `json:"ZipCode"`
	SCID      string `json:"CEEB"`
	SATNumber int    `json:"SATNumber"`
	ACTNumber int    `json:"ACTNumber"`
}
