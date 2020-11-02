package context

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/framework"
	config "github.com/scoir/canis/pkg/mock/config/viper"
)

var endpoint = map[string]interface{}{
	"host": "localhost",
	"port": 8888,
}

func TestProvider_GetStewardClient(t *testing.T) {
	t.Run("test no key", func(t *testing.T) {
		mockConfig := &config.MockConfig{
			EndpointErr: errors.New(""),
		}
		p := NewProvider(mockConfig)

		ep, err := p.GetAPIAdminClient()
		require.Error(t, err)
		require.Nil(t, ep)
	})

	t.Run("test with key", func(t *testing.T) {
		mockConfig := &config.MockConfig{
			EndpointFunc: func(s string) (*framework.Endpoint, error) {
				return &framework.Endpoint{
					Host: "localhost",
					Port: 8888,
				}, nil
			},
		}
		p := NewProvider(mockConfig)

		client, err := p.GetAPIAdminClient()
		require.Nil(t, err)
		require.NotNil(t, client)
	})

}
