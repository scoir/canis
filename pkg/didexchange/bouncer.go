/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didexchange

import (
	"log"
	"time"

	didclient "github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didservice "github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	"github.com/pkg/errors"
)

type Bouncer struct {
	supe  *Supervisor
	didcl *didclient.Client

	validInviteIDs         map[string]bool
	invitationToConnection map[string]string
}

type NotifySuccess func(invitationID string, conn *didclient.Connection)
type NotifyError func(invitationID string, err error)

func NewBouncer(ctx provider) (*Bouncer, error) {
	didcl, err := ctx.GetDIDClient()
	if err != nil {
		return nil, errors.Wrap(err, "error getting did client in bouncer")
	}

	supe, err := New(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error inializing bouncer")
	}

	r := &Bouncer{
		supe:                   supe,
		didcl:                  didcl,
		validInviteIDs:         map[string]bool{},
		invitationToConnection: map[string]string{},
	}

	err = supe.Start(r)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing supervisor in the bouncer")
	}

	return r, nil
}

func (r *Bouncer) InvitationMsg(e didservice.DIDCommAction, invite *didexchange.Invitation) {
	if r.validInviteIDs[invite.ID] {
		e.Continue(didservice.Empty{})
		delete(r.validInviteIDs, invite.ID)
		return
	}

	e.Stop(errors.New("invalid inviteID"))
}

func (r *Bouncer) RequestMsg(e didservice.DIDCommAction, request *didexchange.Request) {
	iID := e.Message.ParentThreadID()
	if r.validInviteIDs[iID] {

		log.Println("received valid request from", request.Connection.DID)
		e.Continue(didservice.Empty{})
		delete(r.validInviteIDs, iID)
		return
	}

	e.Stop(errors.New("invalid parent thread invite ID"))
}

func (r *Bouncer) EstablishConnection(invitation *didclient.Invitation, timeout time.Duration) (*didclient.Connection, error) {
	r.validInviteIDs[invitation.ID] = true
	connectionID, err := r.didcl.HandleInvitation(invitation)
	if err != nil {
		return nil, err
	}

	conn, err := r.waitFor(connectionID, "completed", timeout)
	if err != nil {
		conn, _ = r.didcl.GetConnection(connectionID)
		if conn.State != "completed" {
			return nil, errors.Errorf("connection timed out in bad state: %s", conn.State)
		}
	}

	return conn, nil
}

func (r *Bouncer) CreateInvitation(name string) (*didclient.Invitation, error) {
	invite, err := r.didcl.CreateInvitation(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create invitation in bouncer")
	}
	r.validInviteIDs[invite.ID] = true
	return invite, nil
}

func (r *Bouncer) CreateInvitationNotify(name string, success NotifySuccess, nerr NotifyError) (*didclient.Invitation, error) {
	invitation, err := r.didcl.CreateInvitation(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create invitation in bouncer")
	}
	r.validInviteIDs[invitation.ID] = true

	go func() {
		conn, err := r.waitForInvitation(invitation.ID, "completed")
		if err != nil {
			nerr(invitation.ID, err)
			return
		}
		success(invitation.ID, conn)
	}()

	return invitation, nil
}

func (r *Bouncer) waitFor(connectionID, state string, timeout time.Duration) (*didclient.Connection, error) {
	msgCh := make(chan service.StateMsg)
	_ = r.supe.RegisterMsgEvent(msgCh)
	defer r.Unregister(msgCh)

	for {
		select {
		case e := <-msgCh:
			props, _ := e.Properties.(didclient.Event)
			if props.ConnectionID() == connectionID {
				switch e.StateID {
				case state:
					conn, err := r.didcl.GetConnection(connectionID)
					if err != nil {
						return nil, errors.Wrap(err, "unable to load connection")
					}
					return conn, nil
				case "abandoned":
					return nil, errors.Errorf("connection %s was abandoned", connectionID)
				}
			}
		case <-time.After(timeout):
			log.Println("timeout!!")
			return nil, errors.New("timeout")
		}
	}

}

func (r *Bouncer) waitForInvitation(invitationID, state string) (*didclient.Connection, error) {
	var connectionID string
	msgCh := make(chan service.StateMsg)
	_ = r.supe.RegisterMsgEvent(msgCh)
	defer r.Unregister(msgCh)

	for e := range msgCh {
		props, _ := e.Properties.(didclient.Event)
		if e.Msg.Type() == didexchange.RequestMsgType {
			iID := e.Msg.ParentThreadID()
			if iID == invitationID {
				connectionID = props.ConnectionID()
				continue
			}
		}

		if props.ConnectionID() == connectionID {
			switch e.StateID {
			case state:
				conn, err := r.didcl.GetConnection(connectionID)
				if err != nil {
					return nil, errors.Wrap(err, "unable to load connection")
				}
				return conn, nil
			case "abandoned":
				return nil, errors.Errorf("connection %s was abandoned", connectionID)
			}
		}
	}

	return nil, errors.Errorf("message channel closed for invitation %s", invitationID)
}

func (r *Bouncer) Unregister(ch chan didservice.StateMsg) {
	err := r.supe.UnregisterMsgEvent(ch)
	if err != nil {
		log.Println("error unregistering the bounder state msg channel", err)
	}
}
