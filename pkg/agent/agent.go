/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/pkg/errors"

	ndid "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/steward/api"
	"github.com/scoir/canis/pkg/util"
)

type Agent struct {
	agentID        string
	stewardPeerDID string
	steward        api.AdminClient
	bouncer        ndid.Bouncer
}

type provider interface {
	GetStewardClient() (api.AdminClient, error)
	GetDIDClient() (*didexchange.Client, error)
}

func NewAgent(agentID string, conf provider) (*Agent, error) {
	r := &Agent{
		agentID: agentID,
	}

	var err error
	r.steward, err = conf.GetStewardClient()
	if err != nil {
		return nil, errors.Wrap(err, "error getting steward client for agent")
	}

	r.bouncer, err = ndid.NewBouncer(conf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create did bouncer for agent")
	}

	err = r.bootstrap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to bootstrap agent")
	}

	return r, nil
}

func (r *Agent) bootstrap() error {
	err := backoff.RetryNotify(r.connectToSteward, backoff.NewExponentialBackOff(), util.Logger)
	return errors.Wrap(err, "error connecting to steward in bootstrap")
}

func (r *Agent) connectToSteward() error {
	log.Println("Beginning to connect to steward")
	invite, err := r.steward.GetInvitationForAgent(context.Background(), &api.AgentInvitiationRequest{AgentId: r.agentID})
	if err != nil {
		return errors.Wrap(err, "unable to get invite from steward")
	}

	inv := &didexchange.Invitation{}
	err = json.Unmarshal([]byte(invite.Body), inv)
	if err != nil {
		return errors.Wrap(err, "bad invite from steward")
	}

	log.Println("trying to accept invitation for steward")
	conn, err := r.bouncer.EstablishConnection(inv, 10*time.Second)
	if err != nil {
		return errors.Wrap(err, "unable to establish connection with steward")
	}

	log.Printf("Connected to the Steward with %s\n", conn.TheirDID)
	r.stewardPeerDID = conn.TheirDID

	return nil
}
