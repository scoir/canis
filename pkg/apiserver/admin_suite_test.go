/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiserver

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/scoir/canis/pkg/datastore/mocks"
	dmocks "github.com/scoir/canis/pkg/didexchange/mocks"
)

var target *APIServer

type AdminTestSuite struct {
	suite.Suite
	Store   *mocks.Store
	Bouncer *dmocks.Bouncer
}

func (suite *AdminTestSuite) SetupTest() {
	suite.Store = &mocks.Store{}
	suite.Bouncer = &dmocks.Bouncer{}

	target = &APIServer{
		agentStore:  suite.Store,
		schemaStore: suite.Store,
		didStore:    suite.Store,
	}
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
