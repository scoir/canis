/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didexchange

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	pdid "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	"github.com/pkg/errors"
)

type HandlerFunc func(service.DIDCommAction)
type ConnectionFunc func(didexchange.Connection)

type Handler interface {
	InvitationMsg(e service.DIDCommAction, d *pdid.Invitation)
	RequestMsg(e service.DIDCommAction, d *pdid.Request)
}

type Supervisor struct {
	service.Message
	didcl   service.Event
	actions map[string]chan service.DIDCommAction
}

type provider interface {
	GetDIDClient() (*didexchange.Client, error)
}

func New(ctx provider) (*Supervisor, error) {
	didcl, err := ctx.GetDIDClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create did client in supervisor")
	}

	r := &Supervisor{
		didcl:   didcl,
		actions: make(map[string]chan service.DIDCommAction),
	}

	r.actions[didexchange.InvitationMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[didexchange.RequestMsgType] = make(chan service.DIDCommAction, 1)

	return r, nil
}

func (r *Supervisor) Start(h Handler) error {
	for typ, aCh := range r.actions {
		switch typ {
		case didexchange.InvitationMsgType:
			go r.execInvitationMsg(aCh, h)
		case didexchange.RequestMsgType:
			go r.execRequestMessage(aCh, h)
		}
	}

	aCh := make(chan service.DIDCommAction, 1)
	err := r.didcl.RegisterActionEvent(aCh)
	if err != nil {
		return errors.Wrap(err, "unable to register did action handler in supervisor")
	}

	go r.startActionListener(aCh)
	go r.startMessageListener()

	return nil
}

func (r *Supervisor) startActionListener(aCh chan service.DIDCommAction) {
	for e := range aCh {
		ch, ok := r.actions[e.Message.Type()]
		if ok {
			ch <- e
			continue
		}

		log.Println("unhandled message type in did supervisor:", e.Message.Type())
	}
}

func (r *Supervisor) execInvitationMsg(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		invite := &pdid.Invitation{}
		err := e.Message.Decode(invite)
		if err != nil {
			log.Println("invalid invite object")
		}

		f.InvitationMsg(e, invite)
	}
}
func (r *Supervisor) execRequestMessage(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		request := &pdid.Request{}
		err := e.Message.Decode(request)
		if err != nil {
			log.Println("invalid credential request object")
		}

		f.RequestMsg(e, request)
	}
}

func (r *Supervisor) startMessageListener() {
	didMsgCh := make(chan service.StateMsg)
	_ = r.didcl.RegisterMsgEvent(didMsgCh)
	go func(ch chan service.StateMsg) {
		for msg := range ch {
			if msg.Type == service.PostState {
				for _, c := range r.MsgEvents() {
					c <- msg
				}
				log.Println("DIDEX MSG:", msg.ProtocolName, msg.StateID)
			}
		}
	}(didMsgCh)
}
