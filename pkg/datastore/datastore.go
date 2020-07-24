/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

const (
	SchemaC = "Schema"
	AgentC  = "Agent"
	DIDC    = "PeerDID"
)

// Provider storage provider interface
type Provider interface {
	// OpenStore opens a store with given name space and returns the handle
	OpenStore(name string) (Store, error)

	// CloseStore closes store of given name space
	CloseStore(name string) error

	// Close closes all stores created under this store provider
	Close() error
}

//go:generate mockery -name=Store
type Store interface {
	//InsertDID ..
	InsertDID(d *DID) error
	ListDIDs(c *DIDCriteria) (*DIDList, error)
	SetPublicDID(DID string) error
	GetPublicDID() (*DID, error)

	InsertSchema(s *Schema) (string, error)
	ListSchema(c *SchemaCriteria) (*SchemaList, error)
	GetSchema(id string) (*Schema, error)
	DeleteSchema(id string) error
	UpdateSchema(s *Schema) error

	InsertAgent(s *Agent) (string, error)
	ListAgent(c *AgentCriteria) (*AgentList, error)
	GetAgent(id string) (*Agent, error)
	GetAgentByInvitation(invitationID string) (*Agent, error)
	DeleteAgent(id string) error
	UpdateAgent(s *Agent) error
}
