/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
)

//TODO:  Reset HighSchool / College that was invited - one fail method for each
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

	//TODO: OFFER ALL AVAILABLE CREDENTIALS TO AGENT...  how.  Have to be another API call?
	// err = r.OfferHighSchoolCredential(hs)
	// if err != nil {
	// 	log.Printf("error responding to high school connection activation %s: (%+v)\n", hs.HighSchoolID, err)
	// }
	log.Printf("Agent %s successfully issued Scoir HS credential\n", agent.PeerDID)
}
