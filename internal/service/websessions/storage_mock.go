// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package websessions

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	secret "github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// GetAllForUser provides a mock function with given fields: ctx, userID, cmd
func (_m *MockStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Session, error) {
	ret := _m.Called(ctx, userID, cmd)

	var r0 []Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *sqlstorage.PaginateCmd) ([]Session, error)); ok {
		return rf(ctx, userID, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *sqlstorage.PaginateCmd) []Session); ok {
		r0 = rf(ctx, userID, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *sqlstorage.PaginateCmd) error); ok {
		r1 = rf(ctx, userID, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByToken provides a mock function with given fields: ctx, token
func (_m *MockStorage) GetByToken(ctx context.Context, token secret.Text) (*Session, error) {
	ret := _m.Called(ctx, token)

	var r0 *Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, secret.Text) (*Session, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, secret.Text) *Session); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, secret.Text) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveByToken provides a mock function with given fields: ctx, token
func (_m *MockStorage) RemoveByToken(ctx context.Context, token secret.Text) error {
	ret := _m.Called(ctx, token)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, secret.Text) error); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: ctx, session
func (_m *MockStorage) Save(ctx context.Context, session *Session) error {
	ret := _m.Called(ctx, session)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Session) error); ok {
		r0 = rf(ctx, session)
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
