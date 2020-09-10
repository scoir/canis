package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
)

type MockLoadbalancer struct {
	EndpointValue string
	EndpointErr   error
}

func (r *MockLoadbalancer) GetEndpoint(_ context.Context, _ *api.EndpointRequest, _ ...grpc.CallOption) (*api.EndpointResponse, error) {
	if r.EndpointErr != nil {
		return nil, r.EndpointErr
	}

	return &api.EndpointResponse{Endpoint: r.EndpointValue}, nil
}
