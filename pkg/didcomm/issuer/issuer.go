/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"context"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/issuer/api"
)

type credentialIssuer interface {
	SendOffer(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error)
}

type rpcService struct {
	store                    datastore.Store
	credentialEngineRegistry *engine.Registry
	credHandler              *credHandler
	credcl                   credentialIssuer
}

func NewIssuer(credHandler *credHandler, credcl credentialIssuer, store datastore.Store) *rpcService {
	return &rpcService{
		store:       store,
		credHandler: credHandler,
		credcl:      credcl,
	}
}

func (r *rpcService) IssueCredential(ctx context.Context, req *api.IssueCredentialRequest) (*api.IssueCredentialResponse, error) {

	agent, err := r.store.GetAgent(req.AgentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to load agent: %v", err))
	}

	ac, err := r.store.GetAgentConnection(agent, req.SubjectId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load connection: %v", err))
	}

	schema, err := r.store.GetSchema(req.Credential.SchemaId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to load schema: %v", err))
	}

	attachment, err := r.credentialEngineRegistry.CreateCredentialOffer(agent.PublicDID, schema)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error creating credential offer: %v", err))
	}

	attrs := make([]icprotocol.Attribute, len(req.Credential.Attributes))
	for i, a := range req.Credential.Attributes {
		attrs[i] = icprotocol.Attribute{
			Name:  a.Name,
			Value: a.Value,
		}
	}

	offer := &issuecredential.OfferCredential{
		Comment: req.Credential.Comment,
		CredentialPreview: icprotocol.PreviewCredential{
			Type:       req.Credential.Type,
			Attributes: attrs,
		},
		OffersAttach: []decorator.Attachment{
			{Data: *attachment},
		},
	}

	id, err := r.credcl.SendOffer(offer, ac.MyDID, ac.TheirDID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("send offer failed: %v", err))
	}

	cred := &datastore.Credential{
		AgentID:           agent.ID,
		OfferID:           id,
		ExternalSubjectID: req.SubjectId,
		Offer: datastore.Offer{
			Comment:    req.Credential.Comment,
			Type:       req.Credential.Type,
			Attributes: attrs,
		},
		SystemState: "offered",
	}

	credID, err := r.store.InsertCredential(cred)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unexpected error saving credential: %v", err))
	}

	return &api.IssueCredentialResponse{
		CredentialId: credID,
	}, nil
}
