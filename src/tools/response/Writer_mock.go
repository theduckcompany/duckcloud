// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package response

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// MockWriter is an autogenerated mock type for the Writer type
type MockWriter struct {
	mock.Mock
}

// Write provides a mock function with given fields: w, r, res, statusCode
func (_m *MockWriter) Write(w http.ResponseWriter, r *http.Request, res interface{}, statusCode int) {
	_m.Called(w, r, res, statusCode)
}

// WriteError provides a mock function with given fields: err, w, r
func (_m *MockWriter) WriteError(err error, w http.ResponseWriter, r *http.Request) {
	_m.Called(err, w, r)
}

// NewMockWriter creates a new instance of MockWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWriter {
	mock := &MockWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
