package steward

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/scoir/canis/pkg/framework/context"
	"github.com/scoir/canis/pkg/test"
)

var mongoCfg = map[string]interface{}{
	"database": "mongo",
	"mongo": map[string]interface{}{
		"url":      "mongodb://172.17.0.1:27017",
		"database": "crux",
	},
}

var dockerCfg = map[string]interface{}{
	"runtime": "docker",
	"docker": map[string]interface{}{
		"home": "/etc/canis",
	},
}

func TestNew(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		dir, cleanup := test.GenerateTempDir(t)
		defer cleanup()
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"dbpath":    dir,
			"datastore": mongoCfg,
			"execution": dockerCfg,
		})

		p := context.NewProvider(vp)

		steward, err := New(p)
		assert.Nil(t, err)
		assert.NotNil(t, steward)

	})
}
