package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var endpoint = map[string]interface{}{
	"host": "localhost",
	"port": 8888,
}

func TestProvider_GetGRPCEndpoint(t *testing.T) {
	t.Run("test no key", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		ep, err := p.GetGRPCEndpoint()
		assert.NotNil(t, err)
		assert.Nil(t, ep)
		assert.Equal(t, "grpc is not properly configured", err.Error())
	})
	t.Run("test with key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"grpc": endpoint,
		})
		assert.Nil(t, err)
		p := NewProvider(vp)

		ep, err := p.GetGRPCEndpoint()
		assert.Nil(t, err)
		assert.NotNil(t, ep)
		assert.Equal(t, "localhost:8888", ep.Address())
	})
}

func TestProvider_GetBridgeEndpoint(t *testing.T) {
	t.Run("test no key", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		ep, err := p.GetBridgeEndpoint()
		assert.NotNil(t, err)
		assert.Nil(t, ep)
		assert.Equal(t, "grpc bridge is not properly configured", err.Error())
	})
	t.Run("test with key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"grpcbridge": endpoint,
		})
		assert.Nil(t, err)
		p := NewProvider(vp)

		ep, err := p.GetBridgeEndpoint()
		assert.Nil(t, err)
		assert.NotNil(t, ep)
		assert.Equal(t, "localhost:8888", ep.Address())
	})
}

func TestProvider_GetStewardClient(t *testing.T) {
	t.Run("test no key", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		ep, err := p.GetStewardClient()
		assert.NotNil(t, err)
		assert.Nil(t, ep)
		assert.Equal(t, "steward client is not properly configured", err.Error())
	})
	t.Run("test with key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"stewardEndpoint": endpoint,
		})
		assert.Nil(t, err)
		p := NewProvider(vp)

		client, err := p.GetStewardClient()
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})

}
