package mocks

import (
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
)

type MockVDRClient struct {
	GetNymReply *vdr.ReadReply
	GetNymErr   error
}

func (r *MockVDRClient) CreateNym(did, verkey, role, from string, signer vdr.Signer) error {
	return nil
}

func (r *MockVDRClient) SetEndpoint(did, from string, ep string, signer vdr.Signer) error {
	return nil
}

func (r *MockVDRClient) GetNym(did string) (*vdr.ReadReply, error) {
	if r.GetNymErr != nil {
		return nil, r.GetNymErr
	}

	return r.GetNymReply, nil
}

func (r *MockVDRClient) GetPoolStatus() (*vdr.PoolStatus, error) {
	return &vdr.PoolStatus{}, nil
}
