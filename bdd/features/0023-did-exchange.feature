#
# Copyright Scoir Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
Feature: Decentralized Identifier(DID) exchange between two "always on" agents

  Scenario: did exchange using rest api
    Given "Alice" agent is running on canis with agent id "hogwarts"
      And   "Bob" agent is running on canis with agent id "ministryofmagic"