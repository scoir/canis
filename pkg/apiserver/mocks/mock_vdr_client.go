package mocks

import (
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

type MockVDRClient struct {
}

func (r *MockVDRClient) CreateNym(did, verkey, role, from string, signer vdr.Signer) error {
	return nil
}

func (r *MockVDRClient) SetEndpoint(did, from string, ep string, signer vdr.Signer) error {
	return nil
}

func (r *MockVDRClient) GetNym(did string) (*vdr.ReadReply, error) {
	return &vdr.ReadReply{}, nil
}

func (r *MockVDRClient) GetPoolStatus() (*vdr.PoolStatus, error) {
	return &vdr.PoolStatus{}, nil
}
