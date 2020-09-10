// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	api "github.com/scoir/canis/pkg/didcomm/doorman/api"

	datastore "github.com/scoir/canis/pkg/datastore"

	engine "github.com/scoir/canis/pkg/credential/engine"

	issuerapi "github.com/scoir/canis/pkg/didcomm/issuer/api"

	mock "github.com/stretchr/testify/mock"

	vdr "github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

// Provider is an autogenerated mock type for the provider type
type Provider struct {
	mock.Mock
}

// GetCredentailEngineRegistry provides a mock function with given fields:
func (_m *Provider) GetCredentailEngineRegistry() (engine.CredentialRegistry, error) {
	ret := _m.Called()

	var r0 engine.CredentialRegistry
	if rf, ok := ret.Get(0).(func() engine.CredentialRegistry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(engine.CredentialRegistry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDoormanClient provides a mock function with given fields:
func (_m *Provider) GetDoormanClient() (api.DoormanClient, error) {
	ret := _m.Called()

	var r0 api.DoormanClient
	if rf, ok := ret.Get(0).(func() api.DoormanClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(api.DoormanClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIssuerClient provides a mock function with given fields:
func (_m *Provider) GetIssuerClient() (issuerapi.IssuerClient, error) {
	ret := _m.Called()

	var r0 issuerapi.IssuerClient
	if rf, ok := ret.Get(0).(func() issuerapi.IssuerClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(issuerapi.IssuerClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IndyVDR provides a mock function with given fields:
func (_m *Provider) IndyVDR() (*vdr.Client, error) {
	ret := _m.Called()

	var r0 *vdr.Client
	if rf, ok := ret.Get(0).(func() *vdr.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*vdr.Client)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields:
func (_m *Provider) Store() datastore.Store {
	ret := _m.Called()

	var r0 datastore.Store
	if rf, ok := ret.Get(0).(func() datastore.Store); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.Store)
		}
	}

	return r0
}
