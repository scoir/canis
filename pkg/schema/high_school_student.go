/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

type HighSchoolStudentCredential struct {
	*verifiable.Credential
}

type HighSchoolStudentSubject struct {
	ID             string `json:"id,omitempty"`
	Type           string `json:"type"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	GraduationYear int    `json:"graduation_year"`
}
