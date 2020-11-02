package config

import (
	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/framework"
)

type MockConfig struct {
	EndpointFunc          func(s string) (*framework.Endpoint, error)
	EndpointErr           error
	WithDataStoreFunc     func() config.Config
	WithLedgerStoreFunc   func() config.Config
	WithAMQPFunc          func() config.Config
	WithMasterLockKeyFunc func() config.Config
	WithVDRIFunc          func() config.Config
	WithLedgerGenesisFunc func() config.Config
	WithIndyRegistryFunc  func() config.Config
	MasterLockKeyFunc     func() string
	DataStoreFunc         func() (*framework.DatastoreConfig, error)
	DataStoreErr          error
	VDRIFunc              func() ([]map[string]interface{}, error)
	VDRIErr               error
	LedgerGenesisFunc     func() string
	IndyRegistryFunc      func() string
}

func (m MockConfig) GetInt(s string) int {
	panic("implement me GetInt")
}

func (m MockConfig) WithAMQP(_ ...config.Option) config.Config {
	if m.WithAMQPFunc != nil {
		return m.WithAMQPFunc()
	}

	return nil
}

func (m MockConfig) AMQPAddress() string {
	panic("implement me AMQPAddress")
}

func (m MockConfig) AMQPConfig() (*framework.AMQPConfig, error) {
	panic("implement me AMQPConfig")
}

func (m MockConfig) WithMasterLockKey(_ ...config.Option) config.Config {
	if m.WithMasterLockKeyFunc != nil {
		return m.WithMasterLockKeyFunc()
	}

	return nil
}

func (m MockConfig) MasterLockKey() string {
	if m.MasterLockKeyFunc != nil {
		return m.MasterLockKeyFunc()
	}

	return ""
}

func (m MockConfig) WithDatastore(_ ...config.Option) config.Config {
	if m.WithDataStoreFunc != nil {
		return m.WithDataStoreFunc()
	}

	return nil
}

func (m MockConfig) DataStore() (*framework.DatastoreConfig, error) {
	if m.DataStoreFunc != nil {
		return m.DataStoreFunc()
	}

	if m.DataStoreErr != nil {
		return nil, m.DataStoreErr
	}

	return nil, nil
}

func (m MockConfig) WithLedgerStore(_ ...config.Option) config.Config {
	if m.WithLedgerStoreFunc != nil {
		return m.WithLedgerStoreFunc()
	}

	return nil
}

func (m MockConfig) LedgerStore() (*framework.LedgerStoreConfig, error) {
	panic("implement me LedgerStore")
}

func (m MockConfig) WithVDRI(_ ...config.Option) config.Config {
	if m.WithVDRIFunc != nil {
		return m.WithVDRIFunc()
	}

	return nil
}

func (m MockConfig) VDRIs() ([]map[string]interface{}, error) {
	if m.VDRIFunc != nil {
		return m.VDRIFunc()
	}

	if m.VDRIErr != nil {
		return nil, m.VDRIErr
	}

	return nil, nil
}

func (m MockConfig) GetString(s string) string {
	panic("implement me GetString")
}

func (m MockConfig) Endpoint(s string) (*framework.Endpoint, error) {
	if m.EndpointFunc != nil {
		return m.EndpointFunc(s)
	}

	if m.EndpointErr != nil {
		return nil, m.EndpointErr
	}

	return nil, nil
}

func (m MockConfig) WithLedgerGenesis(opts ...config.Option) config.Config {
	if m.WithLedgerStoreFunc != nil {
		return m.WithLedgerStoreFunc()
	}

	return nil
}

func (m MockConfig) LedgerGenesis() string {
	if m.LedgerGenesisFunc != nil {
		return m.LedgerGenesisFunc()
	}

	return ""
}

func (m MockConfig) WithIndyRegistry(opts ...config.Option) config.Config {
	if m.WithIndyRegistryFunc != nil {
		return m.WithIndyRegistryFunc()
	}

	return nil
}

func (m MockConfig) IndyRegistry() string {
	if m.IndyRegistryFunc != nil {
		return m.IndyRegistryFunc()
	}

	return ""
}
