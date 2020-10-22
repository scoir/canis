package config

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/framework"
)

const (
	defaultAMQP          = "canis-amqp-config"
	defaultVDRI          = "canis-aries-indy-vdri-config"
	defaultDataStore     = "canis-data-store-config"
	defaultLedgerStore   = "canis-ledger-store-config"
	defaultMasterLockKey = "canis-master-lock-key"
	defaultGenesisFile   = "canis-genesis-file"
	defaultIndyRegistry  = "canis-indy-registry"
)

// Option configures the config...
type Option func(opts *vpr)

// WithDBPrefix option is for adding prefix to db name.
func WithFile(file string) Option {
	return func(opts *vpr) {
		opts.file = file
	}
}

type ViperConfigProvider struct {
	DefaultConfigName string
}

type vpr struct {
	*viper.Viper
	file string
}

func (r *ViperConfigProvider) Load(file string) Config {
	config := &vpr{
		viper.New(),
		"", // really don't like this
	}

	if file != "" {
		config.SetConfigFile(file)
	} else {
		config.SetConfigType("yaml")
		config.AddConfigPath("/etc/canis/")
		config.AddConfigPath("./deploy/compose/")
		config.SetConfigName(r.DefaultConfigName)
	}

	config.SetEnvPrefix("CANIS")
	config.AutomaticEnv()

	err := config.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalln("failed to bind flags", err)
	}

	err = config.ReadInConfig()
	if err != nil {
		log.Fatalln("failed to read config after merge", config.ConfigFileUsed(), err)
	}

	return config
}

func (r *vpr) WithVDRI(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultVDRI)
}

func (r *vpr) WithDatastore(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultDataStore)
}

func (r *vpr) WithLedgerStore(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultLedgerStore)
}

func (r *vpr) WithMasterLockKey(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultMasterLockKey)
}

func (r *vpr) WithAMQP(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultAMQP)
}

func (r *vpr) with(file, defawlt string) Config {
	if file != "" {
		return r.withFile(r.SetConfigFile, file)
	}

	return r.withFile(r.SetConfigName, defawlt)
}

func (r *vpr) withFile(setter func(name string), file string) Config {
	setter(file)

	err := r.MergeInConfig()
	if err != nil {
		log.Fatalln("failed to merge", r.ConfigFileUsed(), err)
	}

	return r
}

func (r *vpr) MasterLockKey() string {
	mlk := r.GetString("wallet.masterLockKey")

	if mlk == "" {
		mlk = "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ="
	}

	return mlk
}

func (r *vpr) AMQPAddress() string {
	amqpUser := r.GetString("amqp.user")
	amqpPwd := r.GetString("amqp.password")
	amqpHost := r.GetString("amqp.host")
	amqpPort := r.GetInt("amqp.port")
	amqpVHost := r.GetString("amqp.vhost")

	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", amqpUser, amqpPwd, amqpHost, amqpPort, amqpVHost)
}

func (r *vpr) AMQPConfig() (*framework.AMQPConfig, error) {
	config := &framework.AMQPConfig{}

	err := r.UnmarshalKey("amqp", config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (r *vpr) DataStore() (*framework.DatastoreConfig, error) {
	dc := &framework.DatastoreConfig{}

	err := r.UnmarshalKey("datastore", dc)
	if err != nil {
		return nil, err
	}

	return dc, nil
}

func (r *vpr) LedgerStore() (*framework.LedgerStoreConfig, error) {
	lsc := &framework.LedgerStoreConfig{}

	err := r.UnmarshalKey("ledgerstore", lsc)
	if err != nil {
		return nil, err
	}

	return lsc, nil
}

func (r *vpr) VDRIs() ([]map[string]interface{}, error) {
	var vdri []map[string]interface{}

	err := r.Sub("aries").UnmarshalKey("vdri", &vdri)
	if err != nil {
		return nil, err
	}

	return vdri, nil
}

// GetString uses Get because recursion
func (r *vpr) GetString(s string) string {
	ret, _ := r.Get(s).(string)

	return ret
}

// GetString uses Get because same recursion
func (r *vpr) GetInt(s string) int {
	ret, _ := r.Get(s).(int)

	return ret
}

func (r *vpr) Endpoint(key string) (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}

	err := r.UnmarshalKey(key, ep)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load key "+key)
	}

	return ep, nil
}

func (r *vpr) WithLedgerGenesis(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultGenesisFile)
}

func (r *vpr) LedgerGenesis() string {
	return r.GetString("genesis")
}

func (r *vpr) WithIndyRegistry(opts ...Option) Config {
	for _, opt := range opts {
		opt(r)
	}

	return r.with(r.file, defaultIndyRegistry)
}

func (r *vpr) IndyRegistry() string {
	return r.GetString("registry.indy.genesis")
}
