// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package inodes

import (
	context "context"

	storage "github.com/Peltoche/neurone/src/tools/storage"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/Peltoche/neurone/src/tools/uuid"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// BootstrapUser provides a mock function with given fields: ctx, userID
func (_m *MockService) BootstrapUser(ctx context.Context, userID uuid.UUID) (*INode, error) {
	ret := _m.Called(ctx, userID)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*INode, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *INode); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeletedINodes provides a mock function with given fields: ctx, limit
func (_m *MockService) GetDeletedINodes(ctx context.Context, limit int) ([]INode, error) {
	ret := _m.Called(ctx, limit)

	var r0 []INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) ([]INode, error)); ok {
		return rf(ctx, limit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) []INode); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mkdir provides a mock function with given fields: ctx, cmd
func (_m *MockService) Mkdir(ctx context.Context, cmd *PathCmd) (*INode, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) (*INode, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) *INode); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *PathCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Open provides a mock function with given fields: ctx, cmd
func (_m *MockService) Open(ctx context.Context, cmd *PathCmd) (*INode, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) (*INode, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) *INode); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *PathCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Readdir provides a mock function with given fields: ctx, cmd, paginateCmd
func (_m *MockService) Readdir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error) {
	ret := _m.Called(ctx, cmd, paginateCmd)

	var r0 []INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd, *storage.PaginateCmd) ([]INode, error)); ok {
		return rf(ctx, cmd, paginateCmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd, *storage.PaginateCmd) []INode); ok {
		r0 = rf(ctx, cmd, paginateCmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *PathCmd, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, cmd, paginateCmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveAll provides a mock function with given fields: ctx, cmd
func (_m *MockService) RemoveAll(ctx context.Context, cmd *PathCmd) error {
	ret := _m.Called(ctx, cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) error); ok {
		r0 = rf(ctx, cmd)
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
