// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package password

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockPassword is an autogenerated mock type for the Password type
type MockPassword struct {
	mock.Mock
}

// Compare provides a mock function with given fields: ctx, hash, password
func (_m *MockPassword) Compare(ctx context.Context, hash string, password string) error {
	ret := _m.Called(ctx, hash, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, hash, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Encrypt provides a mock function with given fields: ctx, password
func (_m *MockPassword) Encrypt(ctx context.Context, password string) (string, error) {
	ret := _m.Called(ctx, password)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, password)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, password)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockPassword creates a new instance of MockPassword. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPassword(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPassword {
	mock := &MockPassword{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}