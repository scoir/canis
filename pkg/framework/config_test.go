package framework

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpoint_Address(t *testing.T) {
	t.Run("test address works", func(t *testing.T) {
		ep := Endpoint{
			Host: "localhost",
			Port: 8080,
		}
		addy := ep.Address()
		require.Equal(t, "localhost:8080", addy)
	})
}
