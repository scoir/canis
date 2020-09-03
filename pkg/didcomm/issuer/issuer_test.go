/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"context"
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/didcomm/issuer/api"
)

func TestOfferCredential(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := NewIssuer(&MockDIDClient{
			GetConnectionFunc: func(connectionID string) (*didexchange.Connection, error) {
				return &didexchange.Connection{
					Record: &connection.Record{
						ConnectionID: "connection-id",
						MyDID:        "sample-my-did",
						TheirDID:     "sample-their-did",
					},
				}, nil
			},
		}, &MockIssueCredentialClient{
			VerifySendOfferFunc: func(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error) {
				require.Equal(t, "sample-my-did", myDID)
				require.Equal(t, "sample-their-did", theirDID)
				require.Equal(t, "comment", offer.Comment)
				require.Equal(t, "schema-name", offer.CredentialPreview.Type)

				return "", nil
			},
		})
		require.NotNil(t, s)

		req := &api.OfferCredentialRequest{
			ConnectionID: "connection-id",
			SchemaName:   "schema-name",
			Comment:      "comment",
			Attributes: []*api.CredentialAttribute{
				{Name: "attr", Value: "val"},
			},
		}

		ctx := context.Background()
		_, err := s.OfferCredential(ctx, req)
		require.NoError(t, err)
		//require.NotNil(t, credential)
	})

	t.Run("did client errors", func(t *testing.T) {
		s := NewIssuer(&MockDIDClient{
			GetConnectionErr: errors.New("get connection error"),
		}, &MockIssueCredentialClient{})
		require.NotNil(t, s)

		req := &api.OfferCredentialRequest{
			ConnectionID: "connection-id",
			SchemaName:   "schema-name",
			Comment:      "comment",
		}

		ctx := context.Background()
		_, err := s.OfferCredential(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "get connection error")
	})

	t.Run("issue credential errors", func(t *testing.T) {
		s := NewIssuer(&MockDIDClient{
			GetConnectionFunc: func(connectionID string) (*didexchange.Connection, error) {
				return &didexchange.Connection{
					Record: &connection.Record{
						ConnectionID: "connection-id",
						MyDID:        "sample-my-did",
						TheirDID:     "sample-their-did",
					},
				}, nil
			},
		}, &MockIssueCredentialClient{
			SendOfferErr: errors.New("send offer error"),
		})
		require.NotNil(t, s)

		req := &api.OfferCredentialRequest{
			ConnectionID: "connection-id",
			SchemaName:   "schema-name",
			Comment:      "comment",
		}

		ctx := context.Background()
		_, err := s.OfferCredential(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "send offer error")
	})
}

type MockDIDClient struct {
	GetConnectionErr  error
	GetConnectionFunc func(connectionID string) (*didexchange.Connection, error)
}

func (r *MockDIDClient) GetConnection(connectionID string) (*didexchange.Connection, error) {
	if r.GetConnectionErr != nil {
		return nil, r.GetConnectionErr
	}

	if r.GetConnectionFunc != nil {
		return r.GetConnectionFunc(connectionID)
	}

	return nil, nil
}

type MockIssueCredentialClient struct {
	SendOfferErr        error
	VerifySendOfferFunc func(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error)
}

func (r *MockIssueCredentialClient) SendOffer(offer *issuecredential.OfferCredential, myDID, theirDID string) (string, error) {
	if r.SendOfferErr != nil {
		return "", r.SendOfferErr
	}

	if r.VerifySendOfferFunc != nil {
		return r.VerifySendOfferFunc(offer, myDID, theirDID)
	}

	return "", nil
}
