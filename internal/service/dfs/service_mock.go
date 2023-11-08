// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package dfs

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	folders "github.com/theduckcompany/duckcloud/internal/service/dfs/folders"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// CreateFS provides a mock function with given fields: ctx, owners
func (_m *MockService) CreateFS(ctx context.Context, owners []uuid.UUID) (*folders.Folder, error) {
	ret := _m.Called(ctx, owners)

	var r0 *folders.Folder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []uuid.UUID) (*folders.Folder, error)); ok {
		return rf(ctx, owners)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []uuid.UUID) *folders.Folder); ok {
		r0 = rf(ctx, owners)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*folders.Folder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []uuid.UUID) error); ok {
		r1 = rf(ctx, owners)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFolderFS provides a mock function with given fields: folder
func (_m *MockService) GetFolderFS(folder *folders.Folder) FS {
	ret := _m.Called(folder)

	var r0 FS
	if rf, ok := ret.Get(0).(func(*folders.Folder) FS); ok {
		r0 = rf(folder)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(FS)
		}
	}

	return r0
}

// RemoveFS provides a mock function with given fields: ctx, folder
func (_m *MockService) RemoveFS(ctx context.Context, folder *folders.Folder) error {
	ret := _m.Called(ctx, folder)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *folders.Folder) error); ok {
		r0 = rf(ctx, folder)
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
