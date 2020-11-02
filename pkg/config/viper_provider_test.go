package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestViperConfigProvider_Load(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml")

		require.NotNil(t, conf)
	})
}

func TestVpr_AMQPConfig(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithAMQP(WithFile("./tests/test-amqp-config.yaml"))
		require.NotNil(t, conf)

		cfg, err := conf.AMQPConfig()
		require.NoError(t, err)

		require.Equal(t, "172.17.0.1-test", cfg.Host)
		require.Equal(t, "canis-test", cfg.User)
		require.Equal(t, "canis-test", cfg.Password)
		require.Equal(t, 0, cfg.Port)
		require.Equal(t, "canis-test", cfg.VHost)
	})

	t.Run("bad config", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithAMQP(WithFile("./tests/bad-configs.yaml"))
		require.NotNil(t, conf)

		cfg, err := conf.AMQPConfig()
		require.Error(t, err)
		require.Nil(t, cfg)
	})
}

func TestVpr_AMQPAddress(t *testing.T) {
	vp := &ViperConfigProvider{}
	conf := vp.Load("./tests/test-config.yaml").
		WithAMQP(WithFile("./tests/test-amqp-config.yaml"))
	require.NotNil(t, conf)

	require.Equal(t, "amqp://canis-test:canis-test@172.17.0.1-test:0/canis-test", conf.AMQPAddress())
}

func TestVpr_MasterLockKey(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithMasterLockKey(WithFile("./tests/test-master-lock-key.yaml"))
		require.NotNil(t, conf)

		require.Equal(t, "testy", conf.MasterLockKey())
	})

	t.Run("default key", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithMasterLockKey(WithFile("./tests/bad-configs.yaml"))
		require.NotNil(t, conf)

		require.Equal(t, "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ=", conf.MasterLockKey())
	})
}

func TestVpr_VDRIs(t *testing.T) {
	vp := &ViperConfigProvider{}
	conf := vp.Load("./tests/test-config.yaml").
		WithVDRI(WithFile("./tests/test-aries-vdri-config.yaml"))
	require.NotNil(t, conf)

	vdris, err := conf.VDRIs()
	require.NoError(t, err)
	require.NotNil(t, vdris)

	require.Len(t, vdris, 1)

	require.Equal(t, "scr", vdris[0]["method"])
	require.Equal(t, "indy", vdris[0]["type"])
	require.Equal(t, "genesis-file", vdris[0]["genesisFile"])
}

func TestVpr_DataStore(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithDatastore(WithFile("./tests/test-data-store.yaml"))

		require.NotNil(t, conf)

		ds, err := conf.DataStore()
		require.NoError(t, err)
		require.NotNil(t, ds)

		require.Equal(t, "mongo", ds.Database)
		require.Equal(t, "canis", ds.Mongo.Database)
		require.Equal(t, "mongodb://172.17.0.1:27017", ds.Mongo.URL)
	})

	t.Run("bad config", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithDatastore(WithFile("./tests/bad-configs.yaml"))

		require.NotNil(t, conf)

		ds, err := conf.DataStore()
		require.Error(t, err)
		require.Nil(t, ds)
	})
}

func TestVpr_LedgerStore(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithLedgerStore(WithFile("./tests/test-ledger-store.yaml"))
		require.NotNil(t, conf)

		ls, err := conf.LedgerStore()
		require.NoError(t, err)
		require.NotNil(t, ls)

		require.Equal(t, "mongodb", ls.Database)
		require.Equal(t, "mongodb://172.17.0.1:27017", ls.URL)
	})

	t.Run("bad config", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml").
			WithLedgerStore(WithFile("./tests/bad-configs.yaml"))
		require.NotNil(t, conf)

		ls, err := conf.LedgerStore()
		require.Error(t, err)
		require.Nil(t, ls)
	})
}

func TestVpr_Endpoint(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/test-config.yaml")
		require.NotNil(t, conf)

		ls, err := conf.Endpoint("grpc")
		require.NoError(t, err)
		require.NotNil(t, ls)

		require.Equal(t, "172.17.0.1", ls.Host)
		require.Equal(t, 7776, ls.Port)
		require.Equal(t, "172.17.0.1:7776", ls.Address())
	})

	t.Run("empty end pint", func(t *testing.T) {
		vp := &ViperConfigProvider{}
		conf := vp.Load("./tests/bad-configs.yaml")
		require.NotNil(t, conf)

		ls, err := conf.Endpoint("mis")
		require.NoError(t, err)
		require.NotNil(t, ls)

		require.Equal(t, "", ls.Host)
		require.Equal(t, 0, ls.Port)
	})
}
