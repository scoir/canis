/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/pkg/errors"
)

type HandlerFunc func(service.DIDCommAction)

type Handler interface {
	ProposeCredentialMsg(e service.DIDCommAction, d *icprotocol.ProposeCredential)
	OfferCredentialMsg(e service.DIDCommAction, d *icprotocol.OfferCredential)
	IssueCredentialMsg(e service.DIDCommAction, d *icprotocol.IssueCredential)
	RequestCredentialMsg(e service.DIDCommAction, d *icprotocol.RequestCredential)
}

type Supervisor struct {
	service.Message
	credcli service.Event
	actions map[string]chan service.DIDCommAction
}

type provider interface {
	GetCredentialClient() (*issuecredential.Client, error)
}

func New(ctx provider) (*Supervisor, error) {
	credcli, err := ctx.GetCredentialClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create credential client in supervisor")
	}

	r := &Supervisor{
		credcli: credcli,
		actions: make(map[string]chan service.DIDCommAction),
	}

	r.actions[icprotocol.ProposeCredentialMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[icprotocol.OfferCredentialMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[icprotocol.IssueCredentialMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[icprotocol.RequestCredentialMsgType] = make(chan service.DIDCommAction, 1)

	return r, nil
}

func (r *Supervisor) Start(h Handler) error {
	for typ, aCh := range r.actions {
		switch typ {
		case icprotocol.ProposeCredentialMsgType:
			go r.execProposeCredential(aCh, h)
		case icprotocol.OfferCredentialMsgType:
			go r.execOfferCredential(aCh, h)
		case icprotocol.IssueCredentialMsgType:
			go r.execIssueCredential(aCh, h)
		case icprotocol.RequestCredentialMsgType:
			go r.execRequestCredential(aCh, h)
		}
	}

	aCh := make(chan service.DIDCommAction, 1)
	err := r.credcli.RegisterActionEvent(aCh)
	if err != nil {
		return errors.Wrap(err, "unable to register credential action handler in supervisor")
	}

	go r.startActionListener(aCh)

	return nil
}

func (r *Supervisor) startActionListener(aCh chan service.DIDCommAction) {
	for e := range aCh {
		ch, ok := r.actions[e.Message.Type()]
		if ok {
			ch <- e
			continue
		}

		log.Println("unhandled message type in credential supervisor:", e.Message.Type())
	}
}

func (r *Supervisor) execProposeCredential(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		proposal := &icprotocol.ProposeCredential{}
		err := e.Message.Decode(proposal)
		if err != nil {
			log.Println("invalid credential proposal object")
		}

		f.ProposeCredentialMsg(e, proposal)
	}
}
func (r *Supervisor) execOfferCredential(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		offer := &icprotocol.OfferCredential{}
		err := e.Message.Decode(offer)
		if err != nil {
			log.Println("invalid credential offer object")
		}

		f.OfferCredentialMsg(e, offer)
	}
}
func (r *Supervisor) execIssueCredential(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		issue := &icprotocol.IssueCredential{}
		err := e.Message.Decode(issue)
		if err != nil {
			log.Println("invalid credential issue object")
		}

		f.IssueCredentialMsg(e, issue)
	}
}
func (r *Supervisor) execRequestCredential(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		req := &icprotocol.RequestCredential{}
		err := e.Message.Decode(req)
		if err != nil {
			log.Println("invalid credential req object")
		}

		f.RequestCredentialMsg(e, req)
	}
}
