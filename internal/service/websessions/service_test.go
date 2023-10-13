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
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_WebSessions_Service(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	session := Session{
		token:     "some-token",
		userID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		ip:        "192.168.1.1",
		device:    "Android - Chrome",
		createdAt: now,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
		req.RemoteAddr = "192.168.1.1"

		tools.UUIDMock.On("New").Return(uuid.UUID("some-token")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &session).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			Req:    req,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &session, res)
	})

	t.Run("Create with an invalid cmd", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
		req.RemoteAddr = "192.168.1.1"

		res, err := service.Create(ctx, &CreateCmd{
			UserID: "not a uuid",
			Req:    req,
		})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "UserID: must be a valid UUID v4.")
	})

	t.Run("Create with a storageMockerror", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
		req.RemoteAddr = "192.168.1.1"

		tools.UUIDMock.On("New").Return(uuid.UUID("some-token")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &session).Return(fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			Req:    req,
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "failed to save the session: some-error")
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, AliceWebSessionExample.userID, &storage.PaginateCmd{Limit: 10}).Return([]Session{AliceWebSessionExample}, nil).Once()

		res, err := service.GetAllForUser(ctx, AliceWebSessionExample.userID, &storage.PaginateCmd{Limit: 10})
		assert.NoError(t, err)
		assert.Equal(t, []Session{AliceWebSessionExample}, res)
	})

	t.Run("GetByToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetByToken", mock.Anything, "some-token").Return(&session, nil).Once()

		res, err := service.GetByToken(ctx, "some-token")
		assert.NoError(t, err)
		assert.EqualValues(t, &session, res)
	})

	t.Run("GetFromReq success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storageMock.On("GetByToken", mock.Anything, "some-token").Return(&session, nil).Once()

		res, err := service.GetFromReq(req)
		assert.NoError(t, err)
		assert.EqualValues(t, &session, res)
	})

	t.Run("GetFromReq with no cookie", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		// No cookie

		res, err := service.GetFromReq(req)
		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: missing session token")
	})

	t.Run("GetFromReq with the session not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storageMock.On("GetByToken", mock.Anything, "some-token").Return(nil, nil).Once()

		res, err := service.GetFromReq(req)
		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: session not found")
	})

	t.Run("GetFromReq with a storageMockerror", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storageMock.On("GetByToken", mock.Anything, "some-token").Return(nil, errors.New("some-error")).Once()

		res, err := service.GetFromReq(req)
		assert.Nil(t, res)
		assert.EqualError(t, err, "unhandled: some-error")
	})

	t.Run("Logout success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storageMock.On("RemoveByToken", mock.Anything, "some-token").Return(nil).Once()

		err := service.Logout(req, w)
		assert.NoError(t, err)

		// Check that the session_token cookie is set to an empty value.
		res := w.Result()
		res.Body.Close()
		assert.Len(t, res.Cookies(), 1)
		assert.Empty(t, res.Cookies()[0].Value)
		assert.Equal(t, "session_token", res.Cookies()[0].Name)
	})

	t.Run("Logout with no cookie", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		// No cookie

		// Do nothing

		err := service.Logout(req, w)
		assert.NoError(t, err)
	})

	t.Run("Logout with a storageMockerror", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storageMock.On("RemoveByToken", mock.Anything, "some-token").Return(errors.New("some-error")).Once()

		err := service.Logout(req, w)
		assert.EqualError(t, err, "failed to remove the token: some-error")
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetByToken", mock.Anything, AliceWebSessionExample.Token()).Return(&AliceWebSessionExample, nil).Once()
		storageMock.On("RemoveByToken", mock.Anything, AliceWebSessionExample.Token()).Return(nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID: AliceWebSessionExample.UserID(),
			Token:  AliceWebSessionExample.Token(),
		})
		assert.NoError(t, err)
	})

	t.Run("Delete with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		err := service.Delete(ctx, &DeleteCmd{
			UserID: "some-invalid-id",
			Token:  AliceWebSessionExample.Token(),
		})
		assert.ErrorIs(t, err, errs.ErrValidation)
		assert.ErrorContains(t, err, "UserID: must be a valid UUID v4.")
	})

	t.Run("Delete with a token not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetByToken", mock.Anything, AliceWebSessionExample.Token()).Return(nil, nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID: AliceWebSessionExample.UserID(),
			Token:  AliceWebSessionExample.Token(),
		})
		assert.NoError(t, err)
	})

	t.Run("Delete with a token owned by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetByToken", mock.Anything, AliceWebSessionExample.Token()).Return(&AliceWebSessionExample, nil).Once()

		err := service.Delete(ctx, &DeleteCmd{
			UserID: uuid.UUID("29a81212-9e46-4678-a921-ecaf53aa15bc"), // A random user id
			Token:  AliceWebSessionExample.Token(),
		})
		assert.EqualError(t, err, "not found: user ids are not matching")
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("DeleteAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, AliceWebSessionExample.UserID(), (*storage.PaginateCmd)(nil)).Return([]Session{AliceWebSessionExample}, nil).Once()
		storageMock.On("GetByToken", mock.Anything, AliceWebSessionExample.Token()).Return(&AliceWebSessionExample, nil).Once()
		storageMock.On("RemoveByToken", mock.Anything, AliceWebSessionExample.Token()).Return(nil).Once()

		err := service.DeleteAll(ctx, AliceWebSessionExample.UserID())
		assert.NoError(t, err)
	})

	t.Run("DeleteAll witha GetAll error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, AliceWebSessionExample.UserID(), (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		err := service.DeleteAll(ctx, AliceWebSessionExample.UserID())
		assert.EqualError(t, err, "failed to GetAllForUser: some-error")
	})

	t.Run("DeleteAll with a revoke error stop directly", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, AliceWebSessionExample.UserID(), (*storage.PaginateCmd)(nil)).Return([]Session{AliceWebSessionExample, AliceWebSessionExample}, nil).Once()
		storageMock.On("GetByToken", mock.Anything, AliceWebSessionExample.Token()).Return(&AliceWebSessionExample, nil).Once()
		storageMock.On("RemoveByToken", mock.Anything, AliceWebSessionExample.Token()).Return(fmt.Errorf("some-error")).Once()
		// Do not call GetByToken and RemoveByToken a second time

		err := service.DeleteAll(ctx, AliceWebSessionExample.UserID())
		assert.EqualError(t, err, fmt.Sprintf("failed to Delete web session %q: failed to RemoveByToken: some-error", AliceWebSessionExample.token))
	})
}
