// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package scheduler

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// RegisterFSMoveTask provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterFSMoveTask(ctx context.Context, args *FSMoveArgs) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *FSMoveArgs) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterFSRefreshSizeTask provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterFSRefreshSizeTask(ctx context.Context, args *FSRefreshSizeArg) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *FSRefreshSizeArg) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterFSRemoveDuplicateFile provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterFSRemoveDuplicateFile(ctx context.Context, args *FSRemoveDuplicateFileArgs) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *FSRemoveDuplicateFileArgs) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterFileUploadTask provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterFileUploadTask(ctx context.Context, args *FileUploadArgs) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *FileUploadArgs) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterSpaceCreateTask provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterSpaceCreateTask(ctx context.Context, args *SpaceCreateArgs) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *SpaceCreateArgs) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterUserCreateTask provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterUserCreateTask(ctx context.Context, args *UserCreateArgs) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *UserCreateArgs) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterUserDeleteTask provides a mock function with given fields: ctx, args
func (_m *MockService) RegisterUserDeleteTask(ctx context.Context, args *UserDeleteArgs) error {
	ret := _m.Called(ctx, args)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *UserDeleteArgs) error); ok {
		r0 = rf(ctx, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Run provides a mock function with given fields: ctx
func (_m *MockService) Run(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
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
