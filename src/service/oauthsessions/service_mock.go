// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package oauthsessions

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, input
func (_m *MockService) Create(ctx context.Context, input *CreateCmd) error {
	ret := _m.Called(ctx, input)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *CreateCmd) error); ok {
		r0 = rf(ctx, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByAccessToken provides a mock function with given fields: ctx, access
func (_m *MockService) GetByAccessToken(ctx context.Context, access string) (*Session, error) {
	ret := _m.Called(ctx, access)

	var r0 *Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Session, error)); ok {
		return rf(ctx, access)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Session); ok {
		r0 = rf(ctx, access)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, access)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByRefreshToken provides a mock function with given fields: ctx, refresh
func (_m *MockService) GetByRefreshToken(ctx context.Context, refresh string) (*Session, error) {
	ret := _m.Called(ctx, refresh)

	var r0 *Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Session, error)); ok {
		return rf(ctx, refresh)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Session); ok {
		r0 = rf(ctx, refresh)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, refresh)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveByAccessToken provides a mock function with given fields: ctx, access
func (_m *MockService) RemoveByAccessToken(ctx context.Context, access string) error {
	ret := _m.Called(ctx, access)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, access)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveByRefreshToken provides a mock function with given fields: ctx, refresh
func (_m *MockService) RemoveByRefreshToken(ctx context.Context, refresh string) error {
	ret := _m.Called(ctx, refresh)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, refresh)
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