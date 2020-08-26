/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/pkg/errors"
)

type NotifySuccess func(threadID string, ack *model.Ack)
type NotifyError func(threadID string, err error)

type Monitor struct {
	msgCh chan service.StateMsg
	sup   *Supervisor
}

func NewMonitor(supe *Supervisor) *Monitor {
	r := &Monitor{
		msgCh: make(chan service.StateMsg),
		sup:   supe,
	}

	return r
}

func (r *Monitor) WatchThread(threadID string, success NotifySuccess, nerr NotifyError) {
	go func() {
		ack, err := r.waitForThread(threadID, "done")
		if err != nil {
			nerr(threadID, err)
			return
		}
		success(threadID, ack)
	}()

}

func (r *Monitor) waitForThread(threadID, state string) (*model.Ack, error) {
	_ = r.sup.RegisterMsgEvent(r.msgCh)
	defer r.Unregister()
	for e := range r.msgCh {
		iID, err := e.Msg.ThreadID()
		if err != nil {
			continue
		}
		if iID == threadID {
			switch e.StateID {
			case state:
				out := &model.Ack{}
				err = e.Msg.Decode(out)
				if err != nil {
					return nil, errors.Wrap(err, "unable to decode ack")
				}
				return out, nil
			case "abandoned":
				return nil, errors.New("abandoned")
			}
		}
	}

	return nil, errors.Errorf("message channel closed for credential %s", threadID)
}

func (r *Monitor) Unregister() {
	err := r.sup.UnregisterMsgEvent(r.msgCh)
	if err != nil {
		log.Println("error unregistering credential monitor", err)
	}
}
