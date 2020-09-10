/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"testing"

	"github.com/stretchr/testify/suite"

	emocks "github.com/scoir/canis/pkg/credential/engine/mocks"
	"github.com/scoir/canis/pkg/datastore/mocks"
	dmocks "github.com/scoir/canis/pkg/didexchange/mocks"
)

var target *APIServer

type AdminTestSuite struct {
	suite.Suite
	Store        *mocks.Store
	Bouncer      *dmocks.Bouncer
	CredRegistry *emocks.CredentialRegistry
}

func (suite *AdminTestSuite) SetupTest() {
	suite.Store = &mocks.Store{}
	suite.Bouncer = &dmocks.Bouncer{}
	suite.CredRegistry = &emocks.CredentialRegistry{}

	target = &APIServer{
		agentStore:     suite.Store,
		schemaStore:    suite.Store,
		didStore:       suite.Store,
		schemaRegistry: suite.CredRegistry,
	}
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
