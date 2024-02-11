// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package config

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	secret "github.com/theduckcompany/duckcloud/internal/tools/secret"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// GetMasterKey provides a mock function with given fields: ctx
func (_m *MockService) GetMasterKey(ctx context.Context) (*secret.SealedKey, error) {
	ret := _m.Called(ctx)

	var r0 *secret.SealedKey
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*secret.SealedKey, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *secret.SealedKey); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*secret.SealedKey)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTotalSize provides a mock function with given fields: ctx
func (_m *MockService) GetTotalSize(ctx context.Context) (uint64, error) {
	ret := _m.Called(ctx)

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetMasterKey provides a mock function with given fields: ctx, key
func (_m *MockService) SetMasterKey(ctx context.Context, key *secret.SealedKey) error {
	ret := _m.Called(ctx, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *secret.SealedKey) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTotalSize provides a mock function with given fields: ctx, totalSize
func (_m *MockService) SetTotalSize(ctx context.Context, totalSize uint64) error {
	ret := _m.Called(ctx, totalSize)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64) error); ok {
		r0 = rf(ctx, totalSize)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockService creates a new instance of MockService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockService {
	mock := &MockService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
