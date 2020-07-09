package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
