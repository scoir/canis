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

func TestProvider_GetStewardClient(t *testing.T) {
	t.Run("test no key", func(t *testing.T) {
		vp := viper.New()
		p := NewProvider(vp)

		ep, err := p.GetAPIAdminClient()
		assert.NotNil(t, err)
		assert.Nil(t, ep)
		assert.Equal(t, "api client is not properly configured", err.Error())
	})
	t.Run("test with key", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"api": map[string]interface{}{
				"grpc": endpoint,
			},
		})
		assert.Nil(t, err)
		p := NewProvider(vp)

		client, err := p.GetAPIAdminClient()
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})

}
