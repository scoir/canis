/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package runtime

import (
	"io"
	"time"

	"github.com/scoir/canis/pkg/datastore"
)

type Watcher interface {
	Stop()
	ResultChan() <-chan AgentEvent
}
type Process interface {
	Name() string
	ID() string
	Status() datastore.StatusType
	Config() string
	Time() time.Duration
}

type AgentEvent struct {
	RuntimeContext Process
}

//go:generate mockery -name=Executor
type Executor interface {
	InitSteward(seed string, d []byte) (string, error)
	LaunchSteward([]byte) (string, error)
	ShutdownSteward() error
	LaunchAgent(agent *datastore.Agent) (string, error)
	ShutdownAgent(pID string) error
	AgentPS() []Process

	Status(pID string) (Process, error)
	Watch(pID string) (Watcher, error)
	StreamLogs(pID string) (io.ReadCloser, error)
	PS() []Process
	Describe()
}
