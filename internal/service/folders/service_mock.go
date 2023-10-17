// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package folders

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	storage "github.com/theduckcompany/duckcloud/internal/tools/storage"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// CreatePersonalFolder provides a mock function with given fields: ctx, cmd
func (_m *MockService) CreatePersonalFolder(ctx context.Context, cmd *CreatePersonalFolderCmd) (*Folder, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *Folder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *CreatePersonalFolderCmd) (*Folder, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *CreatePersonalFolderCmd) *Folder); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Folder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *CreatePersonalFolderCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, folderID
func (_m *MockService) Delete(ctx context.Context, folderID uuid.UUID) error {
	ret := _m.Called(ctx, folderID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, folderID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllUserFolders provides a mock function with given fields: ctx, userID, cmd
func (_m *MockService) GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error) {
	ret := _m.Called(ctx, userID, cmd)

	var r0 []Folder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *storage.PaginateCmd) ([]Folder, error)); ok {
		return rf(ctx, userID, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *storage.PaginateCmd) []Folder); ok {
		r0 = rf(ctx, userID, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Folder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, userID, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, folderID
func (_m *MockService) GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error) {
	ret := _m.Called(ctx, folderID)

	var r0 *Folder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*Folder, error)); ok {
		return rf(ctx, folderID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Folder); ok {
		r0 = rf(ctx, folderID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Folder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, folderID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserFolder provides a mock function with given fields: ctx, userID, folderID
func (_m *MockService) GetUserFolder(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) (*Folder, error) {
	ret := _m.Called(ctx, userID, folderID)

	var r0 *Folder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) (*Folder, error)); ok {
		return rf(ctx, userID, folderID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) *Folder); ok {
		r0 = rf(ctx, userID, folderID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Folder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, uuid.UUID) error); ok {
		r1 = rf(ctx, userID, folderID)
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
