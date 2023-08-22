// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package davsessions

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	storage "github.com/theduckcompany/duckcloud/src/tools/storage"

	uuid "github.com/theduckcompany/duckcloud/src/tools/uuid"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// Authenticate provides a mock function with given fields: ctx, username, password
func (_m *MockService) Authenticate(ctx context.Context, username string, password string) (*DavSession, error) {
	ret := _m.Called(ctx, username, password)

	var r0 *DavSession
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*DavSession, error)); ok {
		return rf(ctx, username, password)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *DavSession); ok {
		r0 = rf(ctx, username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*DavSession)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, username, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: ctx, cmd
func (_m *MockService) Create(ctx context.Context, cmd *CreateCmd) (*DavSession, string, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *DavSession
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *CreateCmd) (*DavSession, string, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *CreateCmd) *DavSession); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*DavSession)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *CreateCmd) string); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, *CreateCmd) error); ok {
		r2 = rf(ctx, cmd)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetAllForUser provides a mock function with given fields: ctx, userID, paginateCmd
func (_m *MockService) GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *storage.PaginateCmd) ([]DavSession, error) {
	ret := _m.Called(ctx, userID, paginateCmd)

	var r0 []DavSession
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *storage.PaginateCmd) ([]DavSession, error)); ok {
		return rf(ctx, userID, paginateCmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *storage.PaginateCmd) []DavSession); ok {
		r0 = rf(ctx, userID, paginateCmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]DavSession)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, userID, paginateCmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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
