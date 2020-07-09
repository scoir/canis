package context

import (
	"fmt"
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/route"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/runtime"
)

const (
	agentKey = "agent"
)

type Provider struct {
	vp *viper.Viper

	lock    sync.Mutex
	ds      datastore.Store
	exec    runtime.Executor
	ctx     *context.Provider
	didcl   *didexchange.Client
	credcl  *issuecredential.Client
	routecl *route.Client
}

func NewProvider(vp *viper.Viper) *Provider {
	return &Provider{vp: vp}
}

func (r *Provider) GetAgentConfig(agentID string) (map[string]interface{}, error) {
	agt := r.vp.Sub(agentKey)
	if agt == nil {
		return nil, errors.Errorf("%s is not defined, unable to generate agent config", agentKey)
	}

	dbPathRoot := agt.GetString(dbPathKey)

	ds := r.vp.GetStringMap(datastoreKey)
	agt.Set(datastoreKey, ds)
	agt.Set(dbPathKey, fmt.Sprintf("%s/%s", dbPathRoot, agentID))

	out := map[string]interface{}{}
	err := agt.Unmarshal(&out)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error with unmarshal")
	}
	return out, nil
}

func (r *Provider) GetStewardConfig() (map[string]interface{}, error) {
	ex := r.vp.GetStringMap("execution")
	ds := r.vp.GetStringMap("datastore")
	ag := r.vp.GetStringMap("agent")

	st := r.vp.Sub("steward")
	if st == nil {
		return nil, errors.New("steward key not available")
	}
	st.Set("execution", ex)
	st.Set("datastore", ds)
	st.Set("agent", ag)

	out := map[string]interface{}{}
	err := st.Unmarshal(&out)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected marshal error")
	}

	return out, nil
}
