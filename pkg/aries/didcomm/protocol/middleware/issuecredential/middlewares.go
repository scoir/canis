/*
Copyright Scoir, Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuecredential

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"

	"github.com/scoir/canis/pkg/datastore"
)

const (
	stateNameOfferSent = "offer-sent"
)

// Provider contains dependencies for the SaveCredentials middleware function.
//go:generate mockery -inpkg -name=Provider
type Provider interface {
	Store() datastore.Store
}

// SaveCredentials the helper function for the issue credential protocol which saves credentials.
func SaveCredentials(p Provider) issuecredential.Middleware {
	store := p.Store()

	return func(next issuecredential.Handler) issuecredential.Handler {
		return issuecredential.HandlerFunc(func(metadata issuecredential.Metadata) error {

			state := metadata.StateName()
			if state == stateNameOfferSent {
				return next.Handle(metadata)
			}

			msg := metadata.Message()
			thid, _ := msg.ThreadID()
			cred, err := store.FindCredentialByProtocolID(thid)
			if err != nil {
				return errors.Errorf("unable to find cred with ID %s: (%v)", thid, err)
			}

			cred.SystemState = state

			err = store.UpdateCredential(cred)
			if err != nil {
				return errors.Errorf("unexpected error updating issued credential %s: %v", cred.ID, err)
			}

			return next.Handle(metadata)
		})
	}
}
