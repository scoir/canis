/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
)

const (
	SchemaC = "Schema"
	AgentC  = "Agent"
	DIDC    = "PeerDID"
)

// Provider storage provider interface
//go:generate mockery -name=Provider
type Provider interface {
	// Open opens a store
	Open() (Store, error)

	// Close closes a store
	Close() error
}

//go:generate mockery -name=Store
type Store interface {
	// InsertDID add DID to store
	InsertDID(d *DID) error
	// ListDIDs query DIDs
	ListDIDs(c *DIDCriteria) (*DIDList, error)
	// SetPublicDID update single DID to public, unset remaining
	SetPublicDID(DID *DID) error
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

	// InsertCredential add Crednetial to store
	InsertCredential(c *Credential) (string, error)
	//FindOffer finds credential in offer state
	FindOffer(agentID string, offerID string) (*Credential, error)

	// InsertAgent add agent to store
	InsertAgent(a *Agent) (string, error)
	// ListAgent query agents
	ListAgent(c *AgentCriteria) (*AgentList, error)
	// GetAgent return single agent
	GetAgent(id string) (*Agent, error)
	// GetAgentByIPublicDID return single agent
	GetAgentByPublicDID(DID string) (*Agent, error)
	// DeleteAgent delete single agent
	DeleteAgent(id string) error
	// UpdateAgent delete single agent
	UpdateAgent(a *Agent) error
	// InsertAgentConnection associates an agent with a connection
	InsertAgentConnection(a *Agent, externalID string, conn *didexchange.Connection) error
	// GetAgentConnection return single connection between an agent and an external subject
	GetAgentConnection(a *Agent, externalID string) (*AgentConnection, error)
}
