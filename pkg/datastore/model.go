/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

import "crypto/ed25519"

type Doc interface {
	GetID() string
}

type DocGen func() Doc

type DocList struct {
	Count int
	Docs  []Doc
}

type Criteria map[string]interface{}

type Attribute struct {
	Name string
	Type int32
}

type AgentList struct {
	Count  int
	Agents []*Agent
}

type StatusType string

var (
	NotStarted StatusType = "NOT STARTED"
	Running    StatusType = "RUNNING"
	Error      StatusType = "ERROR"
	Completed  StatusType = "COMPLETED"
)

type Agent struct {
	ID                  string
	Name                string
	AssignedSchemaId    string
	ConnectionID        string
	ConnectionState     string
	PeerDID             string
	EndorsableSchemaIds []string
	Status              StatusType
	PID                 string
	PublicDID           bool
	InvitationID        string
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
	ID         string
	Name       string
	Version    string
	Attributes []*Attribute
}

type Schemas []*Schema

type DIDCriteria struct {
	Start, PageSize int
}

type DIDs []*DID

type DID struct {
	DID, Verkey, Endpoint string
	Public                bool
	Priv                  ed25519.PrivateKey
}

type DIDList struct {
	Count int
	DIDs  []*DID
}
