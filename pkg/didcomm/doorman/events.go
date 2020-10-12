package doorman

type DIDAcceptedEvent struct {
	AgentID      string `json:"agent_id"`
	MyDID        string `json:"my_did"`
	TheirDID     string `json:"their_did"`
	ConnectionID string `json:"connection_id"`
	ExternalID   string `json:"external_id"`
}
