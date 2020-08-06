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
	// InsertDID add DID to store
	InsertDID(d *DID) error
	// ListDIDs query DIDs
	ListDIDs(c *DIDCriteria) (*DIDList, error)
	// SetPublicDID update single DID to public, unset remaining
	SetPublicDID(DID string) error
	// GetPublicDID get public DID
	GetPublicDID() (*DID, error)

	// InsertSchema add Schema to store
	InsertSchema(s *Schema) (string, error)
	// ListSchema query schemas
	ListSchema(c *SchemaCriteria) (*SchemaList, error)
	// GetSchema return single Schema
	GetSchema(id string) (*Schema, error)
	// DeleteSchema delete single schema
	DeleteSchema(id string) error
	// UpdateSchema update single schema
	UpdateSchema(s *Schema) error

	// InsertAgent add agent to store
	InsertAgent(s *Agent) (string, error)
	// ListAgent query agents
	ListAgent(c *AgentCriteria) (*AgentList, error)
	// GetAgent return single agent
	GetAgent(id string) (*Agent, error)
	// GetAgentByInvitation return single agent
	GetAgentByInvitation(invitationID string) (*Agent, error)
	// DeleteAgent delete single agent
	DeleteAgent(id string) error
	// UpdateAgent delete single agent
	UpdateAgent(s *Agent) error
}
