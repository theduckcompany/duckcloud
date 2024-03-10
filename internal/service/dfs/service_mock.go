// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package dfs

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	spaces "github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"

	users "github.com/theduckcompany/duckcloud/internal/service/users"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// CreateDir provides a mock function with given fields: ctx, cmd
func (_m *MockService) CreateDir(ctx context.Context, cmd *CreateDirCmd) (*INode, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *CreateDirCmd) (*INode, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *CreateDirCmd) *INode); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *CreateDirCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateFS provides a mock function with given fields: ctx, user, space
func (_m *MockService) CreateFS(ctx context.Context, user *users.User, space *spaces.Space) (*INode, error) {
	ret := _m.Called(ctx, user, space)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *users.User, *spaces.Space) (*INode, error)); ok {
		return rf(ctx, user, space)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *users.User, *spaces.Space) *INode); ok {
		r0 = rf(ctx, user, space)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *users.User, *spaces.Space) error); ok {
		r1 = rf(ctx, user, space)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Destroy provides a mock function with given fields: ctx, user, space
func (_m *MockService) Destroy(ctx context.Context, user *users.User, space *spaces.Space) error {
	ret := _m.Called(ctx, user, space)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *users.User, *spaces.Space) error); ok {
		r0 = rf(ctx, user, space)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Download provides a mock function with given fields: ctx, cmd
func (_m *MockService) Download(ctx context.Context, cmd *PathCmd) (io.ReadSeekCloser, error) {
	ret := _m.Called(ctx, cmd)

	var r0 io.ReadSeekCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) (io.ReadSeekCloser, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) io.ReadSeekCloser); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadSeekCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *PathCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, cmd
func (_m *MockService) Get(ctx context.Context, cmd *PathCmd) (*INode, error) {
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

// ListDir provides a mock function with given fields: ctx, cmd, paginateCmd
func (_m *MockService) ListDir(ctx context.Context, cmd *PathCmd, paginateCmd *sqlstorage.PaginateCmd) ([]INode, error) {
	ret := _m.Called(ctx, cmd, paginateCmd)

	var r0 []INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd, *sqlstorage.PaginateCmd) ([]INode, error)); ok {
		return rf(ctx, cmd, paginateCmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd, *sqlstorage.PaginateCmd) []INode); ok {
		r0 = rf(ctx, cmd, paginateCmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *PathCmd, *sqlstorage.PaginateCmd) error); ok {
		r1 = rf(ctx, cmd, paginateCmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Move provides a mock function with given fields: ctx, cmd
func (_m *MockService) Move(ctx context.Context, cmd *MoveCmd) error {
	ret := _m.Called(ctx, cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *MoveCmd) error); ok {
		r0 = rf(ctx, cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: ctx, cmd
func (_m *MockService) Remove(ctx context.Context, cmd *PathCmd) error {
	ret := _m.Called(ctx, cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *PathCmd) error); ok {
		r0 = rf(ctx, cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rename provides a mock function with given fields: ctx, inode, newName
func (_m *MockService) Rename(ctx context.Context, inode *INode, newName string) (*INode, error) {
	ret := _m.Called(ctx, inode, newName)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *INode, string) (*INode, error)); ok {
		return rf(ctx, inode, newName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *INode, string) *INode); ok {
		r0 = rf(ctx, inode, newName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *INode, string) error); ok {
		r1 = rf(ctx, inode, newName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Upload provides a mock function with given fields: ctx, cmd
func (_m *MockService) Upload(ctx context.Context, cmd *UploadCmd) error {
	ret := _m.Called(ctx, cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *UploadCmd) error); ok {
		r0 = rf(ctx, cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// removeINode provides a mock function with given fields: ctx, inode
func (_m *MockService) removeINode(ctx context.Context, inode *INode) error {
	ret := _m.Called(ctx, inode)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *INode) error); ok {
		r0 = rf(ctx, inode)
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
