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
	"github.com/scoir/canis/pkg/runtime"
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
		exec    runtime.Executor
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
			r := &Provider{
				vp:      tt.fields.vp,
				lock:    tt.fields.lock,
				exec:    tt.fields.exec,
				ctx:     tt.fields.ctx,
				didcl:   tt.fields.didcl,
				credcl:  tt.fields.credcl,
				routecl: tt.fields.routecl,
			}
			got, err := r.StorageProvider()
			if (err != nil) != tt.wantErr {
				t.Errorf("StorageProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StorageProvider() got = %v, want %v", got, tt.want)
			}
		})
	}
}
