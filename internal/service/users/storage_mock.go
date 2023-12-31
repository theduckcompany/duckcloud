// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package users

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	storage "github.com/theduckcompany/duckcloud/internal/tools/storage"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// GetAll provides a mock function with given fields: ctx, cmd
func (_m *MockStorage) GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]User, error) {
	ret := _m.Called(ctx, cmd)

	var r0 []User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *storage.PaginateCmd) ([]User, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *storage.PaginateCmd) []User); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, userID
func (_m *MockStorage) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	ret := _m.Called(ctx, userID)

	var r0 *User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*User, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *User); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByUsername provides a mock function with given fields: ctx, username
func (_m *MockStorage) GetByUsername(ctx context.Context, username string) (*User, error) {
	ret := _m.Called(ctx, username)

	var r0 *User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*User, error)); ok {
		return rf(ctx, username)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *User); ok {
		r0 = rf(ctx, username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HardDelete provides a mock function with given fields: ctx, userID
func (_m *MockStorage) HardDelete(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Patch provides a mock function with given fields: ctx, userID, fields
func (_m *MockStorage) Patch(ctx context.Context, userID uuid.UUID, fields map[string]interface{}) error {
	ret := _m.Called(ctx, userID, fields)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, map[string]interface{}) error); ok {
		r0 = rf(ctx, userID, fields)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: ctx, user
func (_m *MockStorage) Save(ctx context.Context, user *User) error {
	ret := _m.Called(ctx, user)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *User) error); ok {
		r0 = rf(ctx, user)
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
