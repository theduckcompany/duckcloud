// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package davsessions

import mock "github.com/stretchr/testify/mock"

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
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
