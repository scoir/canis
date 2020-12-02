package mocks

import (
	"context"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/protogen/common"
)

type MockDoorman struct {
	InviteResponse *common.InvitationResponse
	InviteErr      error
}

func (r *MockDoorman) GetInvitation(ctx context.Context, in *common.InvitationRequest, opts ...grpc.CallOption) (*common.InvitationResponse, error) {
	if r.InviteErr != nil {
		return nil, r.InviteErr
	}

	return r.InviteResponse, nil
}

func (r *MockDoorman) AcceptInvitation(ctx context.Context, in *common.AcceptInvitationRequest, opts ...grpc.CallOption) (*common.AcceptInvitationResponse, error) {
	panic("implement me")
}
