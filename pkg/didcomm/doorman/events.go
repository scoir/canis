package doorman

type DIDAcceptedEvent struct {
	AgentName    string `json:"agent_name"`
	MyDID        string `json:"my_did"`
	TheirDID     string `json:"their_did"`
	ConnectionID string `json:"connection_id"`
	ExternalID   string `json:"external_id"`
}
