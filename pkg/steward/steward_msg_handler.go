/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
)

func (r *Steward) failedConnectionHandler(invitationID string, err error) {
	log.Printf("invitation failed for %s: (%+v)\n", invitationID, err)
}

func (r *Steward) handleAgentConnection(invitationID string, conn *didexchange.Connection) {
	agent, err := r.agentStore.GetAgentByInvitation(invitationID)
	if err != nil {
		log.Println(err, "Unable to find for high school %s.", invitationID)
	}

	agent.ConnectionID = conn.ConnectionID
	agent.PeerDID = conn.TheirDID

	agent.ConnectionState = "completed"
	_ = r.agentStore.UpdateAgent(agent)

	log.Printf("Agent %s successfully issued Scoir HS credential\n", agent.PeerDID)
}
