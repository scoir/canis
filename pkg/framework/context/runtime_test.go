package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/scoir/canis/pkg/runtime/docker"
)

var dockerCfg = map[string]interface{}{
	"runtime": "docker",
	"docker": map[string]interface{}{
		"home": "/etc/canis",
	},
}
var badDockerCfg = map[string]interface{}{
	"runtime": "docker",
}
var kubernetesCfg = map[string]interface{}{
	"runtime": "kubernetes",
}

var unknownCfg = map[string]interface{}{
	"runtime": "supervisord",
}

func TestProvider_Executor(t *testing.T) {
	t.Run("test existing env", func(t *testing.T) {
		e := &docker.Executor{}
		p := &Provider{
			exec: e,
		}

		exec, err := p.Executor()
		assert.Nil(t, err)
		assert.Equal(t, e, exec)

	})

	t.Run("test bad doker config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"execution": badDockerCfg,
		})

		assert.Nil(t, err)
		p := &Provider{vp: vp}

		_, err = p.Executor()
		assert.NotNil(t, err)
		assert.Equal(t, "unable to launch runtime from config: docker execution environment not properly configured", err.Error())

	})

	t.Run("test docker config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"execution": dockerCfg,
		})
		assert.Nil(t, err)
		p := &Provider{vp: vp}

		exec, err := p.Executor()
		assert.Nil(t, err)

		_, ok := exec.(*docker.Executor)
		assert.True(t, ok)

	})

	t.Run("test kubernetes config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"execution": kubernetesCfg,
		})

		assert.Nil(t, err)
		p := &Provider{vp: vp}

		_, err = p.Executor()
		assert.NotNil(t, err)
		assert.Equal(t, "unable to launch runtime from config: not implemented", err.Error())

	})

	t.Run("test unkown config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"execution": unknownCfg,
		})

		assert.Nil(t, err)
		p := &Provider{vp: vp}

		_, err = p.Executor()
		assert.NotNil(t, err)
		assert.Equal(t, "no known execution environment is configured", err.Error())
	})
}
