package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/didcomm/issuer/api"
)

type MockIssuer struct {
	IssueCredResponse *api.IssueCredentialResponse
	IssueCredErr      error
}

func (r *MockIssuer) IssueCredential(ctx context.Context, in *api.IssueCredentialRequest, opts ...grpc.CallOption) (*api.IssueCredentialResponse, error) {
	if r.IssueCredErr != nil {
		return nil, r.IssueCredErr
	}

	return r.IssueCredResponse, nil
}
