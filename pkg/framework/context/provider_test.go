package context

import (
	"reflect"
	"sync"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/datastore"
)

var agentCfg = map[string]interface{}{
	"agent": map[string]interface{}{
		"dbpath": "/tmp/path",
	},
}

var agentDsCfg = map[string]interface{}{
	"agent": map[string]interface{}{
		"dbpath": "/tmp/path",
		"datastore": map[string]interface{}{
			"database": "mongo",
		},
	},
}

var stewardCfg = map[string]interface{}{
	"steward": map[string]interface{}{
		"dbpath": "/tmp/path",
	},
}

var stewardFullCfg = map[string]interface{}{
	"steward": map[string]interface{}{
		"dbpath": "/tmp/path",
	},
	"datastore": map[string]interface{}{
		"database": "mongo",
	},
	"execution": map[string]interface{}{
		"runtime": "docker",
	},
	"agent": map[string]interface{}{
		"dbpath": "/tmp/agent",
	},
}

func TestNewProvider(t *testing.T) {
	type args struct {
		vp *viper.Viper
	}
	tests := []struct {
		name string
		args args
		want *Provider
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProvider(tt.args.vp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_StorageProvider(t *testing.T) {
	type fields struct {
		vp      *viper.Viper
		lock    sync.Mutex
		ds      datastore.Provider
		ctx     *context.Provider
		didcl   *didexchange.Client
		credcl  *issuecredential.Client
		routecl *mediator.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    datastore.Provider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//TODO: Same
		})
	}
}
