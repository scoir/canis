/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

//go:generate mockery -name=Store
type Store interface {
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
