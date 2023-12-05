// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package dfs

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	spaces "github.com/theduckcompany/duckcloud/internal/service/spaces"

	storage "github.com/theduckcompany/duckcloud/internal/tools/storage"

	users "github.com/theduckcompany/duckcloud/internal/service/users"
)

// MockFS is an autogenerated mock type for the FS type
type MockFS struct {
	mock.Mock
}

// CreateDir provides a mock function with given fields: ctx, cmd
func (_m *MockFS) CreateDir(ctx context.Context, cmd *CreateDirCmd) (*INode, error) {
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

// Download provides a mock function with given fields: ctx, filePath
func (_m *MockFS) Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error) {
	ret := _m.Called(ctx, filePath)

	var r0 io.ReadSeekCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (io.ReadSeekCloser, error)); ok {
		return rf(ctx, filePath)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) io.ReadSeekCloser); ok {
		r0 = rf(ctx, filePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadSeekCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, cmd
func (_m *MockFS) Get(ctx context.Context, cmd *PathCmd) (*INode, error) {
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

// ListDir provides a mock function with given fields: ctx, dirPath, cmd
func (_m *MockFS) ListDir(ctx context.Context, dirPath string, cmd *storage.PaginateCmd) ([]INode, error) {
	ret := _m.Called(ctx, dirPath, cmd)

	var r0 []INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *storage.PaginateCmd) ([]INode, error)); ok {
		return rf(ctx, dirPath, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *storage.PaginateCmd) []INode); ok {
		r0 = rf(ctx, dirPath, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, dirPath, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Move provides a mock function with given fields: ctx, cmd
func (_m *MockFS) Move(ctx context.Context, cmd *MoveCmd) error {
	ret := _m.Called(ctx, cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *MoveCmd) error); ok {
		r0 = rf(ctx, cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: ctx, path
func (_m *MockFS) Remove(ctx context.Context, path string) error {
	ret := _m.Called(ctx, path)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rename provides a mock function with given fields: ctx, inode, newName
func (_m *MockFS) Rename(ctx context.Context, inode *INode, newName string) (*INode, error) {
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

// Space provides a mock function with given fields:
func (_m *MockFS) Space() *spaces.Space {
	ret := _m.Called()

	var r0 *spaces.Space
	if rf, ok := ret.Get(0).(func() *spaces.Space); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*spaces.Space)
		}
	}

	return r0
}

// Upload provides a mock function with given fields: ctx, cmd
func (_m *MockFS) Upload(ctx context.Context, cmd *UploadCmd) error {
	ret := _m.Called(ctx, cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *UploadCmd) error); ok {
		r0 = rf(ctx, cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// createDir provides a mock function with given fields: ctx, createdBy, parent, name
func (_m *MockFS) createDir(ctx context.Context, createdBy *users.User, parent *INode, name string) (*INode, error) {
	ret := _m.Called(ctx, createdBy, parent, name)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *users.User, *INode, string) (*INode, error)); ok {
		return rf(ctx, createdBy, parent, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *users.User, *INode, string) *INode); ok {
		r0 = rf(ctx, createdBy, parent, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *users.User, *INode, string) error); ok {
		r1 = rf(ctx, createdBy, parent, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// removeINode provides a mock function with given fields: ctx, inode
func (_m *MockFS) removeINode(ctx context.Context, inode *INode) error {
	ret := _m.Called(ctx, inode)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *INode) error); ok {
		r0 = rf(ctx, inode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockFS creates a new instance of MockFS. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFS(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFS {
	mock := &MockFS{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
