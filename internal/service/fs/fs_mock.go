// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package fs

import (
	context "context"
	iofs "io/fs"

	mock "github.com/stretchr/testify/mock"
)

// MockFS is an autogenerated mock type for the FS type
type MockFS struct {
	mock.Mock
}

// CreateDir provides a mock function with given fields: ctx, name
func (_m *MockFS) CreateDir(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Open provides a mock function with given fields: name
func (_m *MockFS) Open(name string) (iofs.File, error) {
	ret := _m.Called(name)

	var r0 iofs.File
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (iofs.File, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) iofs.File); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(iofs.File)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OpenFile provides a mock function with given fields: ctx, name, flag
func (_m *MockFS) OpenFile(ctx context.Context, name string, flag int) (FileOrDirectory, error) {
	ret := _m.Called(ctx, name, flag)

	var r0 FileOrDirectory
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int) (FileOrDirectory, error)); ok {
		return rf(ctx, name, flag)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int) FileOrDirectory); ok {
		r0 = rf(ctx, name, flag)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(FileOrDirectory)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int) error); ok {
		r1 = rf(ctx, name, flag)
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

// Stat provides a mock function with given fields: ctx, name
func (_m *MockFS) Stat(ctx context.Context, name string) (iofs.FileInfo, error) {
	ret := _m.Called(ctx, name)

	var r0 iofs.FileInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (iofs.FileInfo, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) iofs.FileInfo); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(iofs.FileInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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
