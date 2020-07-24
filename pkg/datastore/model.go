/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package datastore

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
}

type AgentCriteria struct {
	Start, PageSize int
	Name            string
}
