package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/protogen/common"
)

type MockMediator struct {
	RegisterResponse *common.RegisterEdgeAgentResponse
	RegisterErr      error
	EndpointResponse *common.EndpointResponse
	EndpointErr      error
}

func (r *MockMediator) RegisterEdgeAgent(ctx context.Context, in *common.RegisterEdgeAgentRequest, opts ...grpc.CallOption) (*common.RegisterEdgeAgentResponse, error) {
	if r.RegisterErr != nil {
		return nil, r.RegisterErr
	}

	return r.RegisterResponse, nil
}

func (r *MockMediator) GetEndpoint(ctx context.Context, in *common.EndpointRequest, opts ...grpc.CallOption) (*common.EndpointResponse, error) {
	if r.EndpointErr != nil {
		return nil, r.EndpointErr
	}

	return r.EndpointResponse, nil
}
