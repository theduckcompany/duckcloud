// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package oauthconsents

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"

	oauthclients "github.com/theduckcompany/duckcloud/internal/service/oauthclients"

	sqlstorage "github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"

	websessions "github.com/theduckcompany/duckcloud/internal/service/websessions"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// Check provides a mock function with given fields: r, client, session
func (_m *MockService) Check(r *http.Request, client *oauthclients.Client, session *websessions.Session) error {
	ret := _m.Called(r, client, session)

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, *oauthclients.Client, *websessions.Session) error); ok {
		r0 = rf(r, client, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: ctx, cmd
func (_m *MockService) Create(ctx context.Context, cmd *CreateCmd) (*Consent, error) {
	ret := _m.Called(ctx, cmd)

	var r0 *Consent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *CreateCmd) (*Consent, error)); ok {
		return rf(ctx, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *CreateCmd) *Consent); ok {
		r0 = rf(ctx, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Consent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *CreateCmd) error); ok {
		r1 = rf(ctx, cmd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, consentID
func (_m *MockService) Delete(ctx context.Context, consentID uuid.UUID) error {
	ret := _m.Called(ctx, consentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, consentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAll provides a mock function with given fields: ctx, userID
func (_m *MockService) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAll provides a mock function with given fields: ctx, userID, cmd
func (_m *MockService) GetAll(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Consent, error) {
	ret := _m.Called(ctx, userID, cmd)

	var r0 []Consent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *sqlstorage.PaginateCmd) ([]Consent, error)); ok {
		return rf(ctx, userID, cmd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *sqlstorage.PaginateCmd) []Consent); ok {
		r0 = rf(ctx, userID, cmd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Consent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *sqlstorage.PaginateCmd) error); ok {
		r1 = rf(ctx, userID, cmd)
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
