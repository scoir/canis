package ursa

import (
	ppprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
)

type Verifier interface {
	SendRequestPresentation(msg *ppprotocol.RequestPresentation, myDID, theirDID string) (string, error)
}

type VerifierServer struct {
}

func NewVerifier() *VerifierServer {
	return &VerifierServer{}
}

func (r *VerifierServer) SendRequestPresentation(msg *ppprotocol.RequestPresentation, myDID, theirDID string) (string, error) {
	return "", nil
}
