package websessions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_WebSessions_Service(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		now := time.Now().UTC()
		rawToken := "some-token"
		user := users.NewFakeUser(t).Build()
		session := NewFakeSession(t).
			WithToken(rawToken).
			WithIP("192.168.1.1").
			WithDevice("Android - Chrome").
			CreatedAt(now).
			CreatedBy(user).
			Build()

		// Mocks
		tools.UUIDMock.On("New").Return(uuid.UUID(rawToken)).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, session).Return(nil).Once()

		// Run
		res, err := service.Create(ctx, &CreateCmd{
			UserID:     user.ID(),
			UserAgent:  "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
			RemoteAddr: "192.168.1.1",
		})

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, session, res)
	})

	t.Run("Create with an invalid cmd", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data

		// Mocks

		// Run
		res, err := service.Create(ctx, &CreateCmd{
			UserID:     "not a uuid",
			UserAgent:  "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
			RemoteAddr: "192.168.1.1",
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "UserID: must be a valid UUID v4.")
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		now := time.Now().UTC()
		rawToken := "some-token"
		user := users.NewFakeUser(t).Build()
		session := NewFakeSession(t).
			WithToken(rawToken).
			WithIP("192.168.1.1").
			WithDevice("Android - Chrome").
			CreatedAt(now).
			CreatedBy(user).
			Build()

		// Mocks
		tools.UUIDMock.On("New").Return(uuid.UUID(rawToken)).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, session).Return(fmt.Errorf("some-error")).Once()

		// Run
		res, err := service.Create(ctx, &CreateCmd{
			UserID:     user.ID(),
			UserAgent:  "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
			RemoteAddr: "192.168.1.1",
		})

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "failed to save the session: some-error")
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		session := NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), &sqlstorage.PaginateCmd{Limit: 10}).Return([]Session{*session}, nil).Once()

		// Run
		res, err := service.GetAllForUser(ctx, user.ID(), &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []Session{*session}, res)
	})

	t.Run("GetByToken success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		rawToken := "some-token"
		session := NewFakeSession(t).WithToken(rawToken).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetByToken", mock.Anything, secret.NewText(rawToken)).Return(session, nil).Once()

		// Run
		res, err := service.GetByToken(ctx, secret.NewText(rawToken))

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, session, res)
	})

	t.Run("GetFromReq success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		rawToken := "some-token"
		session := NewFakeSession(t).WithToken(rawToken).CreatedBy(user).Build()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: rawToken,
		})

		// Mocks
		storageMock.On("GetByToken", mock.Anything, secret.NewText(rawToken)).Return(session, nil).Once()

		// Run
		res, err := service.GetFromReq(req)

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, session, res)
	})

	t.Run("GetFromReq with no cookie", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil) // No cookie

		// Run
		res, err := service.GetFromReq(req)

		// Asserts
		assert.Nil(t, res)
		require.EqualError(t, err, "bad request: missing session token")
	})

	t.Run("GetFromReq with the session not found", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		rawToken := "some-token"
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: rawToken,
		})

		// Mocks
		storageMock.On("GetByToken", mock.Anything, secret.NewText(rawToken)).Return(nil, errNotFound).Once()

		// Run
		res, err := service.GetFromReq(req)

		// Assets
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrBadRequest)
		require.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("GetFromReq with a storage error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		rawToken := "some-token"
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: rawToken,
		})

		// Mocks
		storageMock.On("GetByToken", mock.Anything, secret.NewText(rawToken)).Return(nil, errors.New("some-error")).Once()

		// Run
		res, err := service.GetFromReq(req)

		// Assets
		assert.Nil(t, res)
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("Logout success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		rawToken := "some-token"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: rawToken,
		})

		// Mocks
		storageMock.On("RemoveByToken", mock.Anything, secret.NewText(rawToken)).Return(nil).Once()

		// Run
		err := service.Logout(req, w)

		// Asserts
		require.NoError(t, err)

		res := w.Result() // Check that the session_token cookie is set to an empty value.
		res.Body.Close()
		assert.Len(t, res.Cookies(), 1)
		assert.Empty(t, res.Cookies()[0].Value)
		assert.Equal(t, "session_token", res.Cookies()[0].Name)
	})

	t.Run("Logout with no cookie", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		w := httptest.NewRecorder()

		// Data
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil) // No cookie

		// Mocks
		// Do nothing

		// Run
		err := service.Logout(req, w)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Logout with a storage error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		rawToken := "some-token"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: rawToken,
		})

		// Mocks
		storageMock.On("RemoveByToken", mock.Anything, secret.NewText(rawToken)).Return(errors.New("some-error")).Once()

		// Run
		err := service.Logout(req, w)

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "failed to remove the token: some-error")
	})

	t.Run("Delete success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		rawToken := "some-token"
		session := NewFakeSession(t).WithToken(rawToken).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetByToken", mock.Anything, session.Token()).Return(session, nil).Once()
		storageMock.On("RemoveByToken", mock.Anything, session.Token()).Return(nil).Once()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID: user.ID(),
			Token:  session.Token(),
		})

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Delete with a validation error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		session := NewFakeSession(t).Build()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID: "some-invalid-id",
			Token:  session.Token(),
		})

		// Asserts
		require.ErrorIs(t, err, errs.ErrValidation)
		require.ErrorContains(t, err, "UserID: must be a valid UUID v4.")
	})

	t.Run("Delete with a token not found", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		rawToken := "some-token"
		session := NewFakeSession(t).WithToken(rawToken).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetByToken", mock.Anything, session.Token()).Return(nil, errNotFound).Once()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID: user.ID(),
			Token:  session.Token(),
		})

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Delete with a token owned by someone else", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		anAnotherUser := users.NewFakeUser(t).Build()
		rawToken := "some-token"
		session := NewFakeSession(t).WithToken(rawToken).CreatedBy(anAnotherUser).Build()

		// Mocks
		storageMock.On("GetByToken", mock.Anything, session.Token()).Return(session, nil).Once()

		// Run
		err := service.Delete(ctx, &DeleteCmd{
			UserID: user.ID(),
			Token:  session.Token(), // The sessions is owned by "anAnotherUser"
		})

		// Asserts
		require.EqualError(t, err, "not found: user ids are not matching")
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("DeleteAll success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		rawToken := "some-token"
		session := NewFakeSession(t).WithToken(rawToken).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]Session{*session}, nil).Once()
		storageMock.On("GetByToken", mock.Anything, session.Token()).Return(session, nil).Once()
		storageMock.On("RemoveByToken", mock.Anything, session.Token()).Return(nil).Once()

		// Run
		err := service.DeleteAll(ctx, user.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("DeleteAll with a GetAll error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		// Run
		err := service.DeleteAll(ctx, user.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "failed to GetAllForUser: some-error")
	})

	t.Run("DeleteAll with a revoke error stop directly", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		storageMock := newMockStorage(t)
		service := newService(storageMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()
		session := NewFakeSession(t).CreatedBy(user).Build()
		session2 := NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		storageMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]Session{*session, *session2}, nil).Once()
		storageMock.On("GetByToken", mock.Anything, session.Token()).Return(session, nil).Once()
		storageMock.On("RemoveByToken", mock.Anything, session.Token()).Return(fmt.Errorf("some-error")).Once()
		// Do not call GetByToken and RemoveByToken for "session2"

		// Run
		err := service.DeleteAll(ctx, user.ID())

		// Asserts
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "some-error")
	})
}
