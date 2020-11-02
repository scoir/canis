package issuer

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	CredentialTopic = "credentials"
	ProposedEvent   = "proposed"
)

type CredentialProposalEvent struct {
	AgentID    string                            `json:"agent_id"`
	MyDID      string                            `json:"my_did"`
	TheirDID   string                            `json:"their_did"`
	ExternalID string                            `json:"external_id"`
	Schema     *datastore.Schema                 `json:"schema"`
	Proposal   issuecredential.PreviewCredential `json:"proposal"`
}
