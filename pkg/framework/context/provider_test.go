package context

import (
	"reflect"
	"sync"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/route"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

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

func TestProvider_GetAgentConfig(t *testing.T) {
	t.Run("test no agent key", func(t *testing.T) {
		p := NewProvider(viper.New())
		conf, err := p.GetAgentConfig("123")
		assert.NotNil(t, err)
		assert.Nil(t, conf)
		assert.Equal(t, "agent is not defined, unable to generate agent config", err.Error())
	})

	t.Run("test with agent key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(agentCfg)
		assert.Nil(t, err)

		p := NewProvider(vp)
		conf, err := p.GetAgentConfig("123")
		assert.Nil(t, err)
		assert.Contains(t, conf, "dbpath")
	})

	t.Run("test with datastore key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(agentDsCfg)
		assert.Nil(t, err)

		p := NewProvider(vp)
		conf, err := p.GetAgentConfig("123")
		assert.Nil(t, err)
		assert.Contains(t, conf, "dbpath")
		assert.Contains(t, conf, "datastore")
	})

}

func TestProvider_GetStewardConfig(t *testing.T) {
	t.Run("test no steward key", func(t *testing.T) {
		p := NewProvider(viper.New())
		conf, err := p.GetStewardConfig()
		assert.NotNil(t, err)
		assert.Nil(t, conf)
		assert.Equal(t, "steward key not available", err.Error())
	})

	t.Run("test with steward key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(stewardCfg)
		assert.Nil(t, err)

		p := NewProvider(vp)
		conf, err := p.GetStewardConfig()
		assert.Nil(t, err)
		assert.Contains(t, conf, "dbpath")
	})

	t.Run("test with all keys", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(stewardFullCfg)
		assert.Nil(t, err)

		p := NewProvider(vp)
		conf, err := p.GetStewardConfig()
		assert.Nil(t, err)
		assert.Contains(t, conf, "dbpath")
		assert.Contains(t, conf, "datastore")
		assert.Contains(t, conf, "execution")
		assert.Contains(t, conf, "agent")
	})

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

func TestProvider_GetAgentConfig1(t *testing.T) {
	type fields struct {
		vp      *viper.Viper
		lock    sync.Mutex
		ds      datastore.Provider
		exec    runtime.Executor
		ctx     *context.Provider
		didcl   *didexchange.Client
		credcl  *issuecredential.Client
		routecl *route.Client
	}
	type args struct {
		agentID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Provider{
				vp:      tt.fields.vp,
				lock:    tt.fields.lock,
				ds:      tt.fields.ds,
				exec:    tt.fields.exec,
				ctx:     tt.fields.ctx,
				didcl:   tt.fields.didcl,
				credcl:  tt.fields.credcl,
				routecl: tt.fields.routecl,
			}
			got, err := r.GetAgentConfig(tt.args.agentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAgentConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAgentConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetStewardConfig1(t *testing.T) {
	type fields struct {
		vp      *viper.Viper
		lock    sync.Mutex
		ds      datastore.Provider
		exec    runtime.Executor
		ctx     *context.Provider
		didcl   *didexchange.Client
		credcl  *issuecredential.Client
		routecl *route.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Provider{
				vp:      tt.fields.vp,
				lock:    tt.fields.lock,
				ds:      tt.fields.ds,
				exec:    tt.fields.exec,
				ctx:     tt.fields.ctx,
				didcl:   tt.fields.didcl,
				credcl:  tt.fields.credcl,
				routecl: tt.fields.routecl,
			}
			got, err := r.GetStewardConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStewardConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStewardConfig() got = %v, want %v", got, tt.want)
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
		routecl *route.Client
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
				ds:      tt.fields.ds,
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
