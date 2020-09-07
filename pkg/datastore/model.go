/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

import (
	"github.com/mr-tron/base58"

	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
)

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
	EndorsableSchemaIds []string
	Status              StatusType
	PID                 string
	HasPublicDID        bool
	PublicDID           *DID
}

type AgentConnection struct {
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
	Type             string
	Name             string
	Version          string
	ExternalSchemaID string
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
	PublicKey  string
	PrivateKey string
}

func (r *KeyPair) RawPublicKey() []byte {
	k, _ := base58.Decode(r.PublicKey)
	return k
}

func (r *KeyPair) RawPrivateKey() []byte {
	k, _ := base58.Decode(r.PrivateKey)
	return k
}
