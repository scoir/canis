/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"context"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/didcomm/issuer/api"
)

type didexchanceProvider interface {
	GetConnection(connectionID string) (*didexchange.Connection, error)
}

type issuecredentialProvider interface {
	SendOffer(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error)
}

type rpcService struct {
	api.UnimplementedIssuerServer
	credHandler *credHandler
	credcl      issuecredentialProvider
}

func NewIssuer(credHandler *credHandler, credcl issuecredentialProvider) *rpcService {
	return &rpcService{
		credHandler: credHandler,
		credcl:      credcl,
	}
}

func (r *rpcService) OfferCredential(ctx context.Context, request *api.OfferCredentialRequest) (*api.OfferCredentialResponse, error) {
	if r.credHandler.subject == nil {
		return nil, status.Error(codes.FailedPrecondition, "please connect with the subject first")
	}

	vc := r.credHandler.generateCredential()

	offer := &issuecredential.OfferCredential{
		Comment: "High School Final Transcript",
		CredentialPreview: icprotocol.PreviewCredential{
			Type: "Clr",
			Attributes: []icprotocol.Attribute{
				{
					Name:  "Achievement",
					Value: "Mathmatics - Algebra Level 1",
				},
			},
		},
		OffersAttach: []decorator.Attachment{
			{Data: decorator.AttachmentData{JSON: vc}},
		},
	}

	id, err := r.credcl.SendOffer(offer, r.credHandler.subject.MyDID, r.credHandler.subject.TheirDID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("send offer failed: %v", err))
	}

	r.credHandler.offerID = id

	return nil, nil
}
