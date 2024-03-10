// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package oauthclients

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// mockStorage is an autogenerated mock type for the storage type
type mockStorage struct {
	mock.Mock
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *mockStorage) GetByID(ctx context.Context, id uuid.UUID) (*Client, error) {
	ret := _m.Called(ctx, id)

	var r0 *Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*Client, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Client); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Client)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, client
func (_m *mockStorage) Save(ctx context.Context, client *Client) error {
	ret := _m.Called(ctx, client)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Client) error); ok {
		r0 = rf(ctx, client)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// newMockStorage creates a new instance of mockStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockStorage {
	mock := &mockStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
