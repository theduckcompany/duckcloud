// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package oauthconsents

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	sqlstorage "github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"

	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// mockStorage is an autogenerated mock type for the storage type
type mockStorage struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, consentID
func (_m *mockStorage) Delete(ctx context.Context, consentID uuid.UUID) error {
	ret := _m.Called(ctx, consentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, consentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllForUser provides a mock function with given fields: ctx, userID, cmd
func (_m *mockStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Consent, error) {
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

// GetByID provides a mock function with given fields: ctx, id
func (_m *mockStorage) GetByID(ctx context.Context, id uuid.UUID) (*Consent, error) {
	ret := _m.Called(ctx, id)

	var r0 *Consent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*Consent, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Consent); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Consent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, consent
func (_m *mockStorage) Save(ctx context.Context, consent *Consent) error {
	ret := _m.Called(ctx, consent)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Consent) error); ok {
		r0 = rf(ctx, consent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// newMockStorage creates a new instance of mockStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockStorage {
	mock := &mockStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
