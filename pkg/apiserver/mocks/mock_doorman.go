package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/didcomm/doorman/api"
)

type MockDoorman struct {
	InviteResponse *api.InvitationResponse
	InviteErr      error
}

func (r *MockDoorman) GetInvitation(ctx context.Context, in *api.InvitationRequest, opts ...grpc.CallOption) (*api.InvitationResponse, error) {
	if r.InviteErr != nil {
		return nil, r.InviteErr
	}

	return r.InviteResponse, nil
}
