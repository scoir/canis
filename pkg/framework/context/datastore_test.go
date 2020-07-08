package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/scoir/canis/pkg/datastore/mongodb"
)

var mongoCfg = map[string]interface{}{
	"database": "mongo",
	"mongo": map[string]interface{}{
		"url":      "mongodb://172.17.0.1:27017",
		"database": "crux",
	},
}
var postgresCfg = map[string]interface{}{
	"database": "postgres",
}

func TestProvider_Datastore(t *testing.T) {
	t.Run("test existing store", func(t *testing.T) {
		s := &mongodb.Store{}
		p := &Provider{ds: s}
		store, err := p.Datastore()
		assert.Nil(t, err)
		assert.Equal(t, s, store)
	})

	t.Run("test mongo config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"datastore": mongoCfg,
		})
		assert.Nil(t, err)
		p := &Provider{vp: vp}

		store, err := p.Datastore()
		assert.Nil(t, err)

		_, ok := store.(*mongodb.Store)
		assert.True(t, ok)
	})

	t.Run("no database config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{})
		assert.Nil(t, err)
		p := &Provider{vp: vp}
		store, err := p.Datastore()
		assert.NotNil(t, err)
		assert.Nil(t, store)
	})

	t.Run("postgres", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"datastore": postgresCfg,
		})
		assert.Nil(t, err)
		p := &Provider{vp: vp}

		_, err = p.Datastore()
		assert.NotNil(t, err)
		assert.Equal(t, "unable to get datastore from config: not implemented", err.Error())
	})

}
