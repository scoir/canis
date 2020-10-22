package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/protogen/common"
)

type MockIssuer struct {
	IssueCredResponse *common.IssueCredentialResponse
	IssueCredErr      error
}

func (r *MockIssuer) IssueCredential(ctx context.Context, in *common.IssueCredentialRequest, opts ...grpc.CallOption) (*common.IssueCredentialResponse, error) {
	if r.IssueCredErr != nil {
		return nil, r.IssueCredErr
	}

	return r.IssueCredResponse, nil
}
