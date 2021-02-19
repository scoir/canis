/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

import (
	"time"

	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/mr-tron/base58"
)

type Criteria map[string]interface{}

type Attribute struct {
	Name string
	Type int32
}

type AgentList struct {
	Count  int
	Agents []*Agent
}

type Agent struct {
	ID                    string
	Name                  string
	EndorsableSchemaNames []string
	PID                   string
	HasPublicDID          bool
	PublicDID             *DID
}

func (r *Agent) CanIssue(schemaID string) bool {
	for _, id := range r.EndorsableSchemaNames {
		if schemaID == id {
			return true
		}
	}
	return false
}

type AgentConnection struct {
	TheirLabel   string
	MyLabel      string
	AgentName    string
	TheirDID     string
	MyDID        string
	ConnectionID string
	ExternalID   string
}

type AgentCriteria struct {
	Start, PageSize int
	Name            string
}

type SchemaCriteria struct {
	Start, PageSize int
	Name            string
}

type SchemaList struct {
	Count  int
	Schema []*Schema
}

type Schema struct {
	ID               string
	Format           string
	Type             string
	Name             string
	Version          string
	ExternalSchemaID string
	Context          []string
	Attributes       []*Attribute
}

type Schemas []*Schema

type DIDCriteria struct {
	Start, PageSize int
}

type DIDs []*DID

type DID struct {
	ID       string
	DID      *identifiers.DID
	OwnerID  string
	KeyPair  *KeyPair
	Endpoint string
	Public   bool
}

type DIDList struct {
	Count int
	DIDs  []*DID
}

type KeyPair struct {
	ID        string
	PublicKey string
}

func (r *KeyPair) RawPublicKey() []byte {
	k, _ := base58.Decode(r.PublicKey)
	return k
}

type Offer struct {
	Comment string
	Type    string
	Preview []icprotocol.Attribute
	Data    []byte
}

type Credential struct {
	ID string `json:"@id,omitempty"`
	// Description is an optional human-readable description of the content.
	Description string `json:"description,omitempty"`
	// MimeType describes the MIME type of the attached content. Optional but recommended.
	MimeType string `json:"mime-type,omitempty"`
	// LastModTime is a hint about when the content in this attachment was last modified.
	LastModTime time.Time `json:"lastmod_time,omitempty"`
	Data        []byte    `json:"data"`
}

type IssuedCredential struct {
	ID                string
	AgentName         string
	MyDID             string
	TheirDID          string
	ProtocolID        string
	SchemaName        string
	RegistryOfferID   string
	ExternalSubjectID string
	Offer             *Offer
	Credential        *Credential
	SystemState       string
}

type Webhook struct {
	Type string
	URL  string
}

type PresentationRequest struct {
	AgentID               string
	SchemaID              string
	ExternalID            string
	PresentationRequestID string
	Data                  []byte
}

type Presentation struct {
	TheirDID string
	MyDID    string
	Format   string
	Data     []byte
}

type EdgeAgent struct {
	ID           string
	TheirDID     string
	MyDID        string
	ConnectionID string
	ExternalID   string
}
