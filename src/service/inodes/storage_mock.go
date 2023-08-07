// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package inodes

import (
	context "context"

	uuid "github.com/Peltoche/neurone/src/tools/uuid"
	mock "github.com/stretchr/testify/mock"
)

// MockStorage is an autogenerated mock type for the Storage type
type MockStorage struct {
	mock.Mock
}

// CountUserINodes provides a mock function with given fields: ctx, userID
func (_m *MockStorage) CountUserINodes(ctx context.Context, userID uuid.UUID) (uint, error) {
	ret := _m.Called(ctx, userID)

	var r0 uint
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (uint, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) uint); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(uint)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
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

// GetByNameAndParent provides a mock function with given fields: ctx, userID, name, parent
func (_m *MockStorage) GetByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parent uuid.UUID) (*INode, error) {
	ret := _m.Called(ctx, userID, name, parent)

	var r0 *INode
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, uuid.UUID) (*INode, error)); ok {
		return rf(ctx, userID, name, parent)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, uuid.UUID) *INode); ok {
		r0 = rf(ctx, userID, name, parent)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*INode)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string, uuid.UUID) error); ok {
		r1 = rf(ctx, userID, name, parent)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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
