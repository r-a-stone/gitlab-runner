// Code generated by mockery v1.0.0. DO NOT EDIT.

// This comment works around https://github.com/vektra/mockery/issues/155

package volumes

import context "context"
import mock "github.com/stretchr/testify/mock"

// MockManager is an autogenerated mock type for the Manager type
type MockManager struct {
	mock.Mock
}

// Binds provides a mock function with given fields:
func (_m *MockManager) Binds() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Cleanup provides a mock function with given fields: ctx
func (_m *MockManager) Cleanup(ctx context.Context) chan bool {
	ret := _m.Called(ctx)

	var r0 chan bool
	if rf, ok := ret.Get(0).(func(context.Context) chan bool); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan bool)
		}
	}

	return r0
}

// ContainerIDs provides a mock function with given fields:
func (_m *MockManager) ContainerIDs() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Create provides a mock function with given fields: volume
func (_m *MockManager) Create(volume string) error {
	ret := _m.Called(volume)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(volume)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateTemporary provides a mock function with given fields: containerPath
func (_m *MockManager) CreateTemporary(containerPath string) error {
	ret := _m.Called(containerPath)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(containerPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
