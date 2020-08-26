package context

import (
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/scoir/canis/pkg/test"
)

var baseCfg = map[string]interface{}{
	"dbpath": "/tmp/db",
	"host":   "localhost",
	"port":   8080,
}

var inboundCfg = map[string]interface{}{
	"dbpath": "/tmp/db",
	"host":   "localhost",
	"port":   8080,
	"wsinbound": map[string]interface{}{
		"host": "localhost",
		"port": 8081,
	},
}

func TestProvider_GetAriesContext(t *testing.T) {
	t.Run("get context, no config", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		ctx := p.GetAriesContext()
		assert.NotNil(t, ctx)

		assert.Len(t, ctx.OutboundTransports(), 1)
		assert.Equal(t, "didcomm:transport/queue", ctx.ServiceEndpoint())
		s := ctx.StorageProvider()
		_, ok := s.(*mem.Provider)
		assert.True(t, ok)
	})

	t.Run("get context, base config", func(t *testing.T) {
		d, cleanup := test.GenerateTempDir(t)
		defer cleanup()
		baseCfg["dbpath"] = d
		vp := viper.New()
		err := vp.MergeConfigMap(baseCfg)
		assert.Nil(t, err)
		p := NewProvider(vp)

		ctx := p.GetAriesContext()
		assert.NotNil(t, ctx)

		assert.Len(t, ctx.OutboundTransports(), 1)
		assert.Equal(t, "didcomm:transport/queue", ctx.ServiceEndpoint())
		s := ctx.StorageProvider()
		_, ok := s.(*mem.Provider)
		assert.True(t, ok)
	})

	t.Run("get context, base config, wsinbound", func(t *testing.T) {
		d, cleanup := test.GenerateTempDir(t)
		defer cleanup()
		inboundCfg["dbpath"] = d
		vp := viper.New()
		err := vp.MergeConfigMap(inboundCfg)
		assert.Nil(t, err)
		p := NewProvider(vp)

		ctx := p.GetAriesContext()
		assert.NotNil(t, ctx)

		assert.Len(t, ctx.OutboundTransports(), 1)
		assert.Equal(t, "localhost:8081", ctx.ServiceEndpoint())
		s := ctx.StorageProvider()
		_, ok := s.(*mem.Provider)
		assert.True(t, ok)
	})
}

func TestProvider_GetDIDClient(t *testing.T) {
	t.Run("client already set", func(t *testing.T) {
		dc := &didexchange.Client{}
		p := &Provider{
			didcl: dc,
		}

		didcl, err := p.GetDIDClient()
		assert.Nil(t, err)
		assert.Equal(t, dc, didcl)
	})

	t.Run("client not set", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		didcl, err := p.GetDIDClient()
		assert.Nil(t, err)
		assert.NotNil(t, didcl)
	})
}

func TestProvider_GetSchemaClient(t *testing.T) {
	t.Run("test the stubbed method", func(t *testing.T) {
		p := NewProvider(viper.New())
		sc, err := p.GetSchemaClient()
		assert.Nil(t, err)
		assert.NotNil(t, sc)
	})
}

func TestProvider_GetCredentialClient(t *testing.T) {
	t.Run("client already set", func(t *testing.T) {
		cc := &issuecredential.Client{}
		p := &Provider{
			credcl: cc,
		}

		cccl, err := p.GetCredentialClient()
		assert.Nil(t, err)
		assert.Equal(t, cc, cccl)
	})

	t.Run("client not set", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		cccl, err := p.GetCredentialClient()
		assert.Nil(t, err)
		assert.NotNil(t, cccl)
	})
}

func TestProvider_GetRouterClient(t *testing.T) {
	t.Run("client already set", func(t *testing.T) {
		rc := &mediator.Client{}
		p := &Provider{
			routecl: rc,
		}

		routecl, err := p.GetRouterClient()
		assert.Nil(t, err)
		assert.Equal(t, rc, routecl)
	})

	t.Run("client not set", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		routecl, err := p.GetRouterClient()
		assert.Nil(t, err)
		assert.NotNil(t, routecl)
	})
}

//func TestProvider_newProvider(t *testing.T) {
//	t.Run("initialize with directory", func(t *testing.T) {
//		d, cleanup := test.GenerateTempDir(t)
//		defer cleanup()
//
//		p := newProvider(d)
//		_, ok := p.StorageProvider().(*leveldb.Provider)
//		assert.True(t, ok)
//
//		kms, err := p.createKMS(nil)
//		assert.Nil(t, err)
//		_, ok = kms.(*legacykms.BaseKMS)
//		assert.True(t, ok)
//	})
//}
