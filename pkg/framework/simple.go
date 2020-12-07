package framework

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/client/presentproof"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
)

type SimpleProvider struct {
	ctx *ariescontext.Provider
}

func (r *SimpleProvider) GetPresentProofClient() (*presentproof.Client, error) {
	return presentproof.New(r.ctx)
}

func (r *SimpleProvider) GetCredentialClient() (*issuecredential.Client, error) {
	return issuecredential.New(r.ctx)
}

func NewSimpleProvider(ctx *ariescontext.Provider) *SimpleProvider {
	return &SimpleProvider{ctx: ctx}
}

func (r *SimpleProvider) GetDIDClient() (*didexchange.Client, error) {
	return didexchange.New(r.ctx)
}

func (r *SimpleProvider) GetOOBClient() (*outofband.Client, error) {
	return outofband.New(r.ctx)
}
