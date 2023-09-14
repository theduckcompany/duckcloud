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

// WriteHTML provides a mock function with given fields: w, r, status, template, args
func (_m *MockWriter) WriteHTML(w http.ResponseWriter, r *http.Request, status int, template string, args interface{}) {
	_m.Called(w, r, status, template, args)
}

// WriteHTMLErrorPage provides a mock function with given fields: w, r, err
func (_m *MockWriter) WriteHTMLErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	_m.Called(w, r, err)
}

// WriteJSON provides a mock function with given fields: w, r, statusCode, res
func (_m *MockWriter) WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, res interface{}) {
	_m.Called(w, r, statusCode, res)
}

// WriteJSONError provides a mock function with given fields: w, r, err
func (_m *MockWriter) WriteJSONError(w http.ResponseWriter, r *http.Request, err error) {
	_m.Called(w, r, err)
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
