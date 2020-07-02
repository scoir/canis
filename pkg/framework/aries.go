package framework

import (
	"log"
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/route"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms/legacykms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage/leveldb"
	"github.com/pkg/errors"
)

type AgentConfig struct {
	Endpoint
	DBPath     string   `mapstructure:"dbpath"`
	WSInbound  Endpoint `mapstructure:"wsinbound"`
	GRPC       Endpoint `mapstructure:"grpc"`
	GRPCBridge Endpoint `mapstructure:"grpcbridge"`
	LedgerURL  string   `mapstructure:"ledgerURL"`

	GetAriesOptions func() []aries.Option

	lock    sync.Mutex
	ctx     *context.Provider
	didcl   *didexchange.Client
	credcl  *issuecredential.Client
	routecl *route.Client
}

type provider struct {
	sp  storage.Provider
	kms *legacykms.BaseKMS
}

func newProvider(dbpath string) *provider {
	r := &provider{
		sp: leveldb.NewProvider(dbpath),
	}

	r.kms, _ = legacykms.New(r)
	return r
}

func (r *provider) StorageProvider() storage.Provider {
	return r.sp
}

func (r *provider) createKMS(_ api.Provider) (api.CloseableKMS, error) {
	return r.kms, nil
}

func (r *AgentConfig) GetAriesContext() *context.Provider {
	if r.ctx == nil {
		err := r.createAriesContext()
		if err != nil {
			log.Fatalln("failed to create aries context", err)
		}

	}
	return r.ctx
}

func (r *AgentConfig) createAriesContext() error {
	log.Printf("creating for %s\n", r.DBPath)
	framework, err := aries.New(r.GetAriesOptions()...)
	if err != nil {
		return errors.Wrap(err, "failed to start aries agent rest, failed to initialize framework")
	}

	ctx, err := framework.Context()
	if err != nil {
		return errors.Wrap(err, "failed to start aries agent rest on port, failed to get aries context")
	}
	r.ctx = ctx

	return nil
}

func (r *AgentConfig) GetDIDClient() (*didexchange.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.didcl != nil {
		return r.didcl, nil
	}

	didcl, err := didexchange.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "error creating did client")
	}

	r.didcl = didcl
	return r.didcl, nil
}

func (r *AgentConfig) GetCredentialClient() (*issuecredential.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.credcl != nil {
		return r.credcl, nil
	}

	credcl, err := issuecredential.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "error creating credential client")
	}
	r.credcl = credcl
	return r.credcl, nil
}

func (r *AgentConfig) GetRouterClient() (*route.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.routecl != nil {
		return r.routecl, nil
	}

	routecl, err := route.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create route client for college: %v\n")
	}
	r.routecl = routecl
	return r.routecl, nil
}
