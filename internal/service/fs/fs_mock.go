// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package fs

import (
	context "context"

	folders "github.com/theduckcompany/duckcloud/internal/service/folders"
	inodes "github.com/theduckcompany/duckcloud/internal/service/inodes"

	io "io"

	mock "github.com/stretchr/testify/mock"

	storage "github.com/theduckcompany/duckcloud/internal/tools/storage"
)

// MockFS is an autogenerated mock type for the FS type
type MockFS struct {
	mock.Mock
}

// CreateDir provides a mock function with given fields: ctx, name
func (_m *MockFS) CreateDir(ctx context.Context, name string) (*inodes.INode, error) {
	ret := _m.Called(ctx, name)

	var r0 *inodes.INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*inodes.INode, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *inodes.INode); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*inodes.INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateFile provides a mock function with given fields: ctx, name
func (_m *MockFS) CreateFile(ctx context.Context, name string) (*inodes.INode, error) {
	ret := _m.Called(ctx, name)

	var r0 *inodes.INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*inodes.INode, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *inodes.INode); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*inodes.INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Download provides a mock function with given fields: ctx, inode
func (_m *MockFS) Download(ctx context.Context, inode *inodes.INode) (io.ReadSeekCloser, error) {
	ret := _m.Called(ctx, inode)

	var r0 io.ReadSeekCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *inodes.INode) (io.ReadSeekCloser, error)); ok {
		return rf(ctx, inode)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *inodes.INode) io.ReadSeekCloser); ok {
		r0 = rf(ctx, inode)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadSeekCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *inodes.INode) error); ok {
		r1 = rf(ctx, inode)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Folder provides a mock function with given fields:
func (_m *MockFS) Folder() *folders.Folder {
	ret := _m.Called()

	var r0 *folders.Folder
	if rf, ok := ret.Get(0).(func() *folders.Folder); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*folders.Folder)
		}
	}

	return r0
}

// Get provides a mock function with given fields: ctx, name
func (_m *MockFS) Get(ctx context.Context, name string) (*inodes.INode, error) {
	ret := _m.Called(ctx, name)

	var r0 *inodes.INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*inodes.INode, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *inodes.INode); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*inodes.INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDir provides a mock function with given fields: ctx, name, cmd
func (_m *MockFS) ListDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	ret := _m.Called(ctx, name, cmd)

	var r0 []inodes.INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *storage.PaginateCmd) ([]inodes.INode, error)); ok {
		return rf(ctx, name, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *storage.PaginateCmd) []inodes.INode); ok {
		r0 = rf(ctx, name, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]inodes.INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, name, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveAll provides a mock function with given fields: ctx, name
func (_m *MockFS) RemoveAll(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rename provides a mock function with given fields: ctx, oldName, newName
func (_m *MockFS) Rename(ctx context.Context, oldName string, newName string) error {
	ret := _m.Called(ctx, oldName, newName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, oldName, newName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upload provides a mock function with given fields: ctx, inode, w
func (_m *MockFS) Upload(ctx context.Context, inode *inodes.INode, w io.Reader) error {
	ret := _m.Called(ctx, inode, w)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *inodes.INode, io.Reader) error); ok {
		r0 = rf(ctx, inode, w)
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
