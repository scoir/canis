package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/didcomm/verifier/api"
)

type MockVerifier struct {
	RequestPresResponse *api.RequestPresentationResponse
	RequestPresErr      error
}

func (r *MockVerifier) RequestPresentation(ctx context.Context, in *api.RequestPresentationRequest, opts ...grpc.CallOption) (*api.RequestPresentationResponse, error) {
	if r.RequestPresErr != nil {
		return nil, r.RequestPresErr
	}

	return r.RequestPresResponse, nil
}
