package context

import (
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/spf13/viper"
)

type Provider struct {
	vp *viper.Viper

	lock    sync.Mutex
	ctx     *context.Provider
	didcl   *didexchange.Client
	credcl  *issuecredential.Client
	routecl *mediator.Client
}

func NewProvider(vp *viper.Viper) *Provider {
	return &Provider{vp: vp}
}

func (r *Provider) UnmarshalConfig(dest interface{}) error {
	return r.vp.Unmarshal(dest)
}
