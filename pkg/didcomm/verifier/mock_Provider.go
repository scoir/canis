// Code generated by mockery v1.0.0. DO NOT EDIT.

package verifier

import (
	datastore "github.com/scoir/canis/pkg/datastore"
	engine "github.com/scoir/canis/pkg/presentproof/engine"

	mock "github.com/stretchr/testify/mock"
)

// MockProvider is an autogenerated mock type for the Provider type
type MockProvider struct {
	mock.Mock
}

// GetPresentProofClient provides a mock function with given fields:
func (_m *MockProvider) GetPresentProofClient() (PresentProofClient, error) {
	ret := _m.Called()

	var r0 PresentProofClient
	if rf, ok := ret.Get(0).(func() PresentProofClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(PresentProofClient)
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

// GetPresentationEngineRegistry provides a mock function with given fields:
func (_m *MockProvider) GetPresentationEngineRegistry() (engine.PresentationRegistry, error) {
	ret := _m.Called()

	var r0 engine.PresentationRegistry
	if rf, ok := ret.Get(0).(func() engine.PresentationRegistry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(engine.PresentationRegistry)
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
func (_m *MockProvider) Store() datastore.Store {
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
