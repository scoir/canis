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
	emocks "github.com/scoir/canis/pkg/runtime/mocks"
)

var target *APIServer

type AdminTestSuite struct {
	suite.Suite
	Store   *mocks.Store
	Exec    *emocks.Executor
	Bouncer *dmocks.Bouncer
}

func (suite *AdminTestSuite) SetupTest() {
	suite.Store = &mocks.Store{}
	suite.Exec = &emocks.Executor{}
	suite.Bouncer = &dmocks.Bouncer{}

	target = &APIServer{
		agentStore:  suite.Store,
		schemaStore: suite.Store,
		didStore:    suite.Store,
		exec:        suite.Exec,
	}
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}
