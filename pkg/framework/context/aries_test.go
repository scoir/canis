package context

import (
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/framework"
	mockConfig "github.com/scoir/canis/pkg/mock/config/viper"
)

func TestProvider_GetAriesContext(t *testing.T) {
	t.Run("get context, no config", func(t *testing.T) {
		mc := baseMockConfig()

		p := NewProvider(mc)

		ctx := p.GetAriesContext()
		require.NotNil(t, ctx)

		require.Len(t, ctx.OutboundTransports(), 1)
		require.Equal(t, "didcomm:transport/queue", ctx.ServiceEndpoint())
		s := ctx.StorageProvider()
		_, ok := s.(*mem.Provider)
		require.True(t, ok)
	})
}

func baseMockConfig() *mockConfig.MockConfig {
	mc := &mockConfig.MockConfig{}
	mc.WithDataStoreFunc = func() config.Config {
		return mc
	}
	mc.WithMasterLockKeyFunc = func() config.Config {
		return mc
	}
	mc.WithVDRIFunc = func() config.Config {
		return mc
	}
	mc.MasterLockKeyFunc = func() string {
		return "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ="
	}
	mc.DataStoreFunc = func() (*framework.DatastoreConfig, error) {
		return &framework.DatastoreConfig{
			Database: "",
		}, nil
	}
	mc.VDRIFunc = func() ([]map[string]interface{}, error) {
		var m []map[string]interface{}
		cfgs := map[string]interface{}{
			"aries": map[string]interface{}{
				"type":        "indy",
				"method":      "foo",
				"genesisFile": "file",
			},
		}

		m = append(m, cfgs)
		return m, nil
	}
	return mc
}

func TestProvider_GetDIDClient(t *testing.T) {
	t.Run("client already set", func(t *testing.T) {
		dc := &didexchange.Client{}
		p := &Provider{
			didcl: dc,
		}

		didcl, err := p.GetDIDClient()
		require.NoError(t, err)
		require.Equal(t, dc, didcl)
	})

	t.Run("client not set", func(t *testing.T) {

		p := NewProvider(baseMockConfig())

		didcl, err := p.GetDIDClient()
		require.NoError(t, err)
		require.NotNil(t, didcl)
	})
}

func TestProvider_GetCredentialClient(t *testing.T) {
	t.Run("client already set", func(t *testing.T) {
		cc := &issuecredential.Client{}
		p := &Provider{
			credcl: cc,
		}

		cccl, err := p.GetCredentialClient()
		require.NoError(t, err)
		require.Equal(t, cc, cccl)
	})

	t.Run("client not set", func(t *testing.T) {
		p := NewProvider(baseMockConfig())

		cccl, err := p.GetCredentialClient()
		require.NoError(t, err)
		require.NotNil(t, cccl)
	})
}
