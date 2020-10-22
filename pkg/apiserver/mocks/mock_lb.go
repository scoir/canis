package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/protogen/common"
)

type MockLoadbalancer struct {
	EndpointValue string
	EndpointErr   error
}

func (r *MockLoadbalancer) GetEndpoint(_ context.Context, _ *common.EndpointRequest, _ ...grpc.CallOption) (*common.EndpointResponse, error) {
	if r.EndpointErr != nil {
		return nil, r.EndpointErr
	}

	return &common.EndpointResponse{Endpoint: r.EndpointValue}, nil
}
