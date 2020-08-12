package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/datastore/mongodb"
)

func TestProvider_Datastore(t *testing.T) {

	var mongoCfg = map[string]interface{}{
		"database": "mongo",
		"mongo": map[string]interface{}{
			"url":      "mongodb://172.17.0.1:27017",
			"database": "crux",
		},
	}

	t.Run("test mongo config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"datastore": mongoCfg,
		})
		require.NoError(t, err)
		p := &Provider{vp: vp}

		store, err := p.Datastore()
		require.NoError(t, err)

		_, ok := store.(*mongodb.Provider)
		require.True(t, ok)
	})

	t.Run("no database config", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{})
		require.NoError(t, err)

		p := &Provider{vp: vp}
		store, err := p.Datastore()
		require.Error(t, err)
		require.Nil(t, store)
	})
}

func TestProvider_PostgresDatastore(t *testing.T) {
	var postgresCfg = map[string]interface{}{
		"database": "postgres",
		"postgres": map[string]interface{}{
			"database": "canis",
			"host":     "127.0.0.1",
			"port":     5432,
			"user":     "postgres",
			"password": "mysecretpassword",
			"sslmode":  "disable",
		},
	}

	t.Run("postgres", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"datastore": postgresCfg,
		})
		require.NoError(t, err)
		p := &Provider{vp: vp}

		ds, err := p.Datastore()
		require.NoError(t, err)
		require.NotNil(t, ds)
	})
}

func TestProvider_CouchDBDatastore(t *testing.T) {
	var couchdbCfg = map[string]interface{}{
		"database": "couchdb",
		"couchdb": map[string]interface{}{
			"url": "localhost:5984",
		},
	}

	t.Run("postgres", func(t *testing.T) {
		vp := viper.New()
		err := vp.MergeConfigMap(map[string]interface{}{
			"datastore": couchdbCfg,
		})
		require.NoError(t, err)
		p := &Provider{vp: vp}

		ds, err := p.Datastore()
		require.NoError(t, err)
		require.NotNil(t, ds)
	})
}
