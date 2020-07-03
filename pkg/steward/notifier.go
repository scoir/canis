/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/webnotifier"
	"github.com/pkg/errors"
	goji "goji.io"
	"goji.io/pat"
)

func (r *Steward) registerWebNotifier(mux *goji.Mux) {
	r.notifier = webnotifier.New("/steward/notify", []string{})
	handlers := r.notifier.GetRESTHandlers()
	handler := handlers[0]
	mux.Handle(pat.NewWithMethods(handler.Path(), handler.Method()), handler.Handle())
}

func (r *Steward) sendMsg(topic string, msg interface{}) error {
	d, err := json.MarshalIndent(msg, " ", "")
	if err != nil {
		return errors.Wrap(err, "unable to marshal notification")
	}

	err = r.notifier.Notify(topic, d)
	return errors.Wrap(err, "error sending WS notification")
}
