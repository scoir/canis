/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

import (
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/mr-tron/base58"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
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
	AgentID      string
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
	Body    []byte
}

type Credential struct {
	AgentID           string
	MyDID             string
	TheirDID          string
	ThreadID          string
	SchemaID          string
	RegistryOfferID   string
	ExternalSubjectID string
	Offer             Offer
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
}
