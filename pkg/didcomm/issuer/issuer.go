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
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"

	"github.com/scoir/canis/pkg/didcomm/issuer/api"
)

type didexchanceProvider interface {
	GetConnection(connectionID string) (*didexchange.Connection, error)
}

type issuecredentialProvider interface {
	SendOffer(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error)
}

type IssuerService struct {
	api.UnimplementedIssuerServer
	didcl  didexchanceProvider
	credcl issuecredentialProvider
}

func NewIssuer(didcl didexchanceProvider, credcl issuecredentialProvider) *IssuerService {
	return &IssuerService{
		didcl:  didcl,
		credcl: credcl,
	}
}

func (s IssuerService) OfferCredential(ctx context.Context, request *api.OfferCredentialRequest) (*api.OfferCredentialResponse, error) {
	conn, err := s.didcl.GetConnection(request.ConnectionID)
	if err != nil {
		return nil, err
	}

	var attrs []icprotocol.Attribute
	for _, v := range request.Attributes {
		attrs = append(attrs, icprotocol.Attribute{
			Name:  v.Name,
			Value: v.Value,
		})
	}

	offer := &issuecredential.OfferCredential{
		Comment: request.Comment,
		CredentialPreview: icprotocol.PreviewCredential{
			Type:       request.SchemaName,
			Attributes: attrs,
		},
		OffersAttach: nil,
	}

	sss, err := s.credcl.SendOffer(offer, conn.MyDID, conn.TheirDID)
	if err != nil {
		return nil, err
	}

	fmt.Println(sss)

	return nil, nil
}
