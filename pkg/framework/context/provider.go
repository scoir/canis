package context

import (
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/scoir/canis/pkg/config"
)

type Provider struct {
	conf    config.Config
	lock    sync.Mutex
	ctx     *context.Provider
	didcl   *didexchange.Client
	credcl  *issuecredential.Client
	routecl *mediator.Client
}

func NewProvider(conf config.Config) *Provider {
	return &Provider{conf: conf}
}
