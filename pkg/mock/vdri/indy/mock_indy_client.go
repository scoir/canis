package indy

import (
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

type MockIndyClient struct {
	GetNymErr      error
	GetNymValue    *vdr.ReadReply
	GetEndpointErr error
	GetEndpointVal *vdr.ReadReply
	RefreshErr     error
	CloseErr       error
}

func (r *MockIndyClient) GetNym(did string) (*vdr.ReadReply, error) {
	if r.GetNymErr != nil {
		return nil, r.GetNymErr
	}

	return r.GetNymValue, nil
}

func (r *MockIndyClient) GetEndpoint(did string) (*vdr.ReadReply, error) {
	if r.GetEndpointErr != nil {
		return nil, r.GetEndpointErr
	}

	return r.GetEndpointVal, nil
}

func (r *MockIndyClient) RefreshPool() error {
	return r.RefreshErr
}

func (r *MockIndyClient) Close() error {
	return r.CloseErr
}

func (r *MockIndyClient) GetPoolStatus() (*vdr.PoolStatus, error) {
	return &vdr.PoolStatus{}, nil
}
