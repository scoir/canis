/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
)

func (r *Steward) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {
}

func (r *Steward) OfferCredentialMsg(e service.DIDCommAction, d *icprotocol.OfferCredential) {
}

func (r *Steward) IssueCredentialMsg(e service.DIDCommAction, d *icprotocol.IssueCredential) {
}

func (r *Steward) RequestCredentialMsg(e service.DIDCommAction, request *icprotocol.RequestCredential) {
}
