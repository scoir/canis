package config

import "github.com/scoir/canis/pkg/framework"

// Provider rename to ConfigBuilder
type Provider interface {
	Load(file string) Config
}

// Config
type Config interface {
	WithAMQP(opts ...Option) Config
	AMQPAddress() string
	AMQPConfig() (*framework.AMQPConfig, error)

	WithMasterLockKey(opts ...Option) Config
	MasterLockKey() string

	WithDatastore(opts ...Option) Config
	DataStore() (*framework.DatastoreConfig, error)

	WithLedgerStore(opts ...Option) Config
	LedgerStore() (*framework.LedgerStoreConfig, error)

	WithVDRI(opts ...Option) Config
	VDRIs() ([]map[string]interface{}, error)

	GetString(s string) string
	GetInt(s string) int

	Endpoint(s string) (*framework.Endpoint, error)

	WithLedgerGenesis(opts ...Option) Config
	LedgerGenesis() string

	WithIndyRegistry(opts ...Option) Config
	IndyRegistry() string
}
