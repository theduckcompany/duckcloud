// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package jwt

import (
	http "net/http"

	generates "github.com/go-oauth2/oauth2/v4/generates"

	mock "github.com/stretchr/testify/mock"
)

// MockParser is an autogenerated mock type for the Parser type
type MockParser struct {
	mock.Mock
}

// FetchAccessToken provides a mock function with given fields: r, permissions
func (_m *MockParser) FetchAccessToken(r *http.Request, permissions ...string) (*AccessToken, error) {
	_va := make([]interface{}, len(permissions))
	for _i := range permissions {
		_va[_i] = permissions[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, r)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *AccessToken
	var r1 error
	if rf, ok := ret.Get(0).(func(*http.Request, ...string) (*AccessToken, error)); ok {
		return rf(r, permissions...)
	}
	if rf, ok := ret.Get(0).(func(*http.Request, ...string) *AccessToken); ok {
		r0 = rf(r, permissions...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*AccessToken)
		}
	}

	if rf, ok := ret.Get(1).(func(*http.Request, ...string) error); ok {
		r1 = rf(r, permissions...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateAccess provides a mock function with given fields:
func (_m *MockParser) GenerateAccess() *generates.JWTAccessGenerate {
	ret := _m.Called()

	var r0 *generates.JWTAccessGenerate
	if rf, ok := ret.Get(0).(func() *generates.JWTAccessGenerate); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*generates.JWTAccessGenerate)
		}
	}

	return r0
}

// NewMockParser creates a new instance of MockParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockParser(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockParser {
	mock := &MockParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}