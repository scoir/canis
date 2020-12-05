/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import "github.com/cucumber/godog"

func agentIsRunningOnCanisWithAgentId(arg1, arg2 string) error {
	//make rest call
	return godog.ErrPending
}

func FeatureContext(s *godog.ScenarioContext) {
	s.Step(`^"([^"]*)" agent is running on canis with agent id "([^"]*)"$`, agentIsRunningOnCanisWithAgentId)
}
