/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
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
	GetDID(id string) (*DID, error)
	// SetPublicDID update single DID to public, unset remaining
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
	GetSchema(name string) (*Schema, error)
	// GetSchema return single Schema
	GetSchemaByExternalID(externalID string) (*Schema, error)
	// DeleteSchema delete single schema
	DeleteSchema(name string) error
	// UpdateSchema update single schema
	UpdateSchema(s *Schema) error

	// InsertCredential add Crednetial to store
	InsertCredential(c *IssuedCredential) (string, error)
	//FindCredentialByOffer finds credential in offer state
	FindCredentialByProtocolID(offerID string) (*IssuedCredential, error)
	// Update Credentia updates the credential
	UpdateCredential(c *IssuedCredential) error
	//Delete offer deletes the offer identifed by the offerID (thid of message)
	DeleteCredentialByOffer(offerID string) error

	// InsertAgent add agent to store
	InsertAgent(a *Agent) (string, error)
	// ListAgent query agents
	ListAgent(c *AgentCriteria) (*AgentList, error)
	// GetAgent return single agent
	GetAgent(id string) (*Agent, error)
	// GetAgentByIPublicDID return single agent
	GetAgentByPublicDID(DID string) (*Agent, error)
	// DeleteAgent delete single agent
	DeleteAgent(name string) error
	// UpdateAgent delete single agent
	UpdateAgent(a *Agent) error
	// InsertAgentConnection associates an agent with a connection
	InsertAgentConnection(a *Agent, externalID string, conn *didexchange.Connection) error
	// ListAgentConnections returns all the connections for an agent
	ListAgentConnections(a *Agent) ([]*AgentConnection, error)
	// GetAgentConnection return single connection between an agent and an external subject
	GetAgentConnection(a *Agent, externalID string) (*AgentConnection, error)
	// DeleteAgentConnection deletes a connection for an agent
	DeleteAgentConnection(a *Agent, externalID string) error

	// GetAgentConnectionForDID return single connection between an agent and an external subject
	GetAgentConnectionForDID(a *Agent, theirDID string) (*AgentConnection, error)

	// ListWebhooks list the webhooks currently registered
	ListWebhooks(typ string) ([]*Webhook, error)

	// AddWebhook adds a new webhook
	AddWebhook(hook *Webhook) error

	// DeleteWebhook deletes an existing webhook
	DeleteWebhook(typ string) error

	//InsertPresentationRequest inserts the presentation request
	InsertPresentationRequest(pr *PresentationRequest) (string, error)

	// InsertPresentation inserts the presentation
	InsertPresentation(p *Presentation) (string, error)

	// GetPresentationRequest retrieves the presentation request by ID
	GetPresentationRequest(ID string) (*PresentationRequest, error)

	// SetMediatorDID sets the public DID for the mediator
	SetMediatorDID(DID *DID) error

	// GetMediatorDID get public DID for the mediator
	GetMediatorDID() (*DID, error)

	// RegisterEdgeAgent associates the DID and external ID with an internal ID for a registered
	// Edge Agent
	RegisterEdgeAgent(connectionID, externalID string) (string, error)

	// GetEdgeAgent retrieves the internal ID associated with a registered DID for an
	// Edge Agent
	GetEdgeAgent(connectionID string) (*EdgeAgent, error)

	// GetEdgeAgentForDID retrieves the internal ID associated with a registered DID for an existing connection
	GetEdgeAgentForDID(theirDID string) (*EdgeAgent, error)

	// UpdateEdgeAgent updates the edge agent using the connection ID and the external ID
	UpdateEdgeAgent(ea *EdgeAgent) error

	// RegisterCloudAgent associates the DID and external ID with an internal ID for a registered
	// Cloud Agent
	RegisterCloudAgent(externalID string, publicKey, nextKey []byte) (string, error)

	// GetCloudAgent retrieves the internal ID associated with a registered DID for an
	// Cloud Agent
	GetCloudAgent(ID string) (*CloudAgent, error)

	// GetCloudAgentForDID retrieves the internal ID associated with a registered DID for an existing connection
	GetCloudAgentForDID(theirDID string) (*CloudAgent, error)

	// UpdateCloudAgent updates the cloud agent using the connection ID and the external ID
	UpdateCloudAgent(ea *CloudAgent) error

	// InsertAgentConnection associates an agent with a connection
	InsertCloudAgentConnection(a *CloudAgent, conn *didexchange.Connection) error
	// ListAgentConnections returns all the connections for an agent
	ListCloudAgentConnections(a *CloudAgent) ([]*CloudAgentConnection, error)
	// GetAgentConnection return single connection between an agent and an external subject
	GetCloudAgentConnection(a *CloudAgent, externalID string) (*CloudAgentConnection, error)
	// DeleteAgentConnection deletes a connection for an agent
	DeleteCloudAgentConnection(a *CloudAgent, externalID string) error

	// GetAgentConnectionForDID return single connection between an agent and an external subject
	GetCloudAgentConnectionForDID(a *CloudAgent, theirDID string) (*CloudAgentConnection, error)
}
