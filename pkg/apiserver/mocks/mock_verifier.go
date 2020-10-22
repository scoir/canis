package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/protogen/common"
)

type MockVerifier struct {
	RequestPresResponse *common.RequestPresentationResponse
	RequestPresErr      error
}

func (r *MockVerifier) RequestPresentation(ctx context.Context, in *common.RequestPresentationRequest, opts ...grpc.CallOption) (*common.RequestPresentationResponse, error) {
	if r.RequestPresErr != nil {
		return nil, r.RequestPresErr
	}

	return r.RequestPresResponse, nil
}
