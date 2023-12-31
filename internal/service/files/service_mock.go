// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package files

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, fileID
func (_m *MockService) Delete(ctx context.Context, fileID uuid.UUID) error {
	ret := _m.Called(ctx, fileID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, fileID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Download provides a mock function with given fields: ctx, file
func (_m *MockService) Download(ctx context.Context, file *FileMeta) (io.ReadSeekCloser, error) {
	ret := _m.Called(ctx, file)

	var r0 io.ReadSeekCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *FileMeta) (io.ReadSeekCloser, error)); ok {
		return rf(ctx, file)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *FileMeta) io.ReadSeekCloser); ok {
		r0 = rf(ctx, file)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadSeekCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *FileMeta) error); ok {
		r1 = rf(ctx, file)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMetadata provides a mock function with given fields: ctx, fileID
func (_m *MockService) GetMetadata(ctx context.Context, fileID uuid.UUID) (*FileMeta, error) {
	ret := _m.Called(ctx, fileID)

	var r0 *FileMeta
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*FileMeta, error)); ok {
		return rf(ctx, fileID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *FileMeta); ok {
		r0 = rf(ctx, fileID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*FileMeta)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, fileID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Upload provides a mock function with given fields: ctx, r
func (_m *MockService) Upload(ctx context.Context, r io.Reader) (*FileMeta, error) {
	ret := _m.Called(ctx, r)

	var r0 *FileMeta
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) (*FileMeta, error)); ok {
		return rf(ctx, r)
	}
	if rf, ok := ret.Get(0).(func(context.Context, io.Reader) *FileMeta); ok {
		r0 = rf(ctx, r)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*FileMeta)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, io.Reader) error); ok {
		r1 = rf(ctx, r)
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
