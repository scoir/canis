package context

import (
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/runtime"
)

const (
	agentKey = "agent"
)

type Provider struct {
	vp *viper.Viper

	lock         sync.Mutex
	datastoreMgr manager.DataProviderManager
	exec         runtime.Executor
	ctx          *context.Provider
	didcl        *didexchange.Client
	credcl       *issuecredential.Client
	routecl      *mediator.Client
}

func NewProvider(vp *viper.Viper) *Provider {
	return &Provider{vp: vp}
}

func (r *Provider) UnmarshalConfig(dest interface{}) error {
	return r.vp.Unmarshal(dest)
}

func (r *Provider) StorageProvider() (datastore.Provider, error) {
	return r.Datastore()
}
