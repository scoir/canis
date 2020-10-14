/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package presentproof

import (
	"log"

	ppclient "github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/pkg/errors"
)

type HandlerFunc func(service.DIDCommAction)

type Handler interface {
	ProposePresentationMsg(e service.DIDCommAction, d *presentproof.ProposePresentation)
	RequestPresentationMsg(e service.DIDCommAction, d *presentproof.RequestPresentation)
	PresentationMsg(e service.DIDCommAction, d *presentproof.Presentation)
	PresentationPreviewMsg(e service.DIDCommAction, d *presentproof.Presentation)
}

type Supervisor struct {
	service.Message
	ppcli   service.Event
	actions map[string]chan service.DIDCommAction
}

type provider interface {
	GetPresentProofClient() (*ppclient.Client, error)
}

func New(ctx provider) (*Supervisor, error) {
	ppcli, err := ctx.GetPresentProofClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create present proof client in supervisor")
	}

	r := &Supervisor{
		ppcli:   ppcli,
		actions: make(map[string]chan service.DIDCommAction),
	}

	r.actions[presentproof.ProposePresentationMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[presentproof.RequestPresentationMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[presentproof.PresentationMsgType] = make(chan service.DIDCommAction, 1)
	r.actions[presentproof.PresentationPreviewMsgType] = make(chan service.DIDCommAction, 1)

	return r, nil
}

func (r *Supervisor) Start(h Handler) error {
	for typ, aCh := range r.actions {
		switch typ {
		case presentproof.ProposePresentationMsgType:
			go r.execProposePresentation(aCh, h)
		case presentproof.RequestPresentationMsgType:
			go r.execRequestPresentation(aCh, h)
		case presentproof.PresentationMsgType:
			go r.execPresentation(aCh, h)
		case presentproof.PresentationPreviewMsgType:
			go r.execPresentationPreview(aCh, h)
		}
	}

	aCh := make(chan service.DIDCommAction, 1)
	err := r.ppcli.RegisterActionEvent(aCh)
	if err != nil {
		return errors.Wrap(err, "unable to register present proof action handler in supervisor")
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

		log.Println("unhandled message type in proof supervisor:", e.Message.Type())
	}
}

func (r *Supervisor) execProposePresentation(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		proposal := &presentproof.ProposePresentation{}
		err := e.Message.Decode(proposal)
		if err != nil {
			log.Println("invalid proposal object")
		}

		f.ProposePresentationMsg(e, proposal)
	}
}

func (r *Supervisor) execRequestPresentation(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		request := &presentproof.RequestPresentation{}
		err := e.Message.Decode(request)
		if err != nil {
			log.Println("invalid presentation request object")
		}

		f.RequestPresentationMsg(e, request)
	}
}

func (r *Supervisor) execPresentation(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		pres := &presentproof.Presentation{}
		err := e.Message.Decode(pres)
		if err != nil {
			log.Println("invalid presentation object")
		}

		f.PresentationMsg(e, pres)
	}
}

func (r *Supervisor) execPresentationPreview(ch chan service.DIDCommAction, f Handler) {
	for e := range ch {
		pres := &presentproof.Presentation{}
		err := e.Message.Decode(pres)
		if err != nil {
			log.Println("invalid proof presentation object")
		}

		f.PresentationPreviewMsg(e, pres)
	}
}

func (r *Supervisor) startMessageListener() {
	proofMsgCh := make(chan service.StateMsg)
	_ = r.ppcli.RegisterMsgEvent(proofMsgCh)
	go func(ch chan service.StateMsg) {
		for msg := range ch {
			if msg.Type == service.PostState {
				for _, c := range r.MsgEvents() {
					c <- msg
				}
				thid, _ := msg.Msg.ThreadID()
				pthid := msg.Msg.ParentThreadID()
				log.Println("PROOF MSG:", msg.ProtocolName, msg.StateID, thid, pthid)
			}
		}
	}(proofMsgCh)
}
