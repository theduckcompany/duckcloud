// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package oauthconsents

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *MockStorage) GetByID(ctx context.Context, id string) (*Consent, error) {
	ret := _m.Called(ctx, id)

	var r0 *Consent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Consent, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Consent); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Consent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, consent
func (_m *MockStorage) Save(ctx context.Context, consent *Consent) error {
	ret := _m.Called(ctx, consent)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Consent) error); ok {
		r0 = rf(ctx, consent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockStorage creates a new instance of MockStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStorage {
	mock := &MockStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
