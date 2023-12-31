// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package dfs

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	storage "github.com/theduckcompany/duckcloud/internal/tools/storage"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// GetAllChildrens provides a mock function with given fields: ctx, parent, cmd
func (_m *MockStorage) GetAllChildrens(ctx context.Context, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error) {
	ret := _m.Called(ctx, parent, cmd)

	var r0 []INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *storage.PaginateCmd) ([]INode, error)); ok {
		return rf(ctx, parent, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *storage.PaginateCmd) []INode); ok {
		r0 = rf(ctx, parent, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *storage.PaginateCmd) error); ok {
		r1 = rf(ctx, parent, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllDeleted provides a mock function with given fields: ctx, limit
func (_m *MockStorage) GetAllDeleted(ctx context.Context, limit int) ([]INode, error) {
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

// GetAllInodesWithFileID provides a mock function with given fields: ctx, fileID
func (_m *MockStorage) GetAllInodesWithFileID(ctx context.Context, fileID uuid.UUID) ([]INode, error) {
	ret := _m.Called(ctx, fileID)

	var r0 []INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]INode, error)); ok {
		return rf(ctx, fileID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []INode); ok {
		r0 = rf(ctx, fileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, fileID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *MockStorage) GetByID(ctx context.Context, id uuid.UUID) (*INode, error) {
	ret := _m.Called(ctx, id)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*INode, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *INode); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByNameAndParent provides a mock function with given fields: ctx, name, parent
func (_m *MockStorage) GetByNameAndParent(ctx context.Context, name string, parent uuid.UUID) (*INode, error) {
	ret := _m.Called(ctx, name, parent)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) (*INode, error)); ok {
		return rf(ctx, name, parent)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) *INode); ok {
		r0 = rf(ctx, name, parent)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, uuid.UUID) error); ok {
		r1 = rf(ctx, name, parent)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeleted provides a mock function with given fields: ctx, id
func (_m *MockStorage) GetDeleted(ctx context.Context, id uuid.UUID) (*INode, error) {
	ret := _m.Called(ctx, id)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*INode, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *INode); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSpaceRoot provides a mock function with given fields: ctx, spaceID
func (_m *MockStorage) GetSpaceRoot(ctx context.Context, spaceID uuid.UUID) (*INode, error) {
	ret := _m.Called(ctx, spaceID)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*INode, error)); ok {
		return rf(ctx, spaceID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *INode); ok {
		r0 = rf(ctx, spaceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, spaceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSumChildsSize provides a mock function with given fields: ctx, parent
func (_m *MockStorage) GetSumChildsSize(ctx context.Context, parent uuid.UUID) (uint64, error) {
	ret := _m.Called(ctx, parent)

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (uint64, error)); ok {
		return rf(ctx, parent)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) uint64); ok {
		r0 = rf(ctx, parent)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, parent)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HardDelete provides a mock function with given fields: ctx, id
func (_m *MockStorage) HardDelete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Patch provides a mock function with given fields: ctx, inode, fields
func (_m *MockStorage) Patch(ctx context.Context, inode uuid.UUID, fields map[string]interface{}) error {
	ret := _m.Called(ctx, inode, fields)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, map[string]interface{}) error); ok {
		r0 = rf(ctx, inode, fields)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: ctx, dir
func (_m *MockStorage) Save(ctx context.Context, dir *INode) error {
	ret := _m.Called(ctx, dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *INode) error); ok {
		r0 = rf(ctx, dir)
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
