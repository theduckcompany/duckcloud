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
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_WebSessions_Service(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	session := Session{
		token:     "some-token",
		userID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		ip:        "192.168.1.1",
		clientID:  "some-client-id",
		device:    "Android - Chrome",
		createdAt: now,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
		req.RemoteAddr = "192.168.1.1"

		tools.UUIDMock.On("New").Return(uuid.UUID("some-token")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storage.On("Save", mock.Anything, &session).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID:   "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID: "some-client-id",
			Req:      req,
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &session, res)
	})

	t.Run("Create with an invalid cmd", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
		req.RemoteAddr = "192.168.1.1"

		res, err := service.Create(ctx, &CreateCmd{
			UserID:   "not a uuid",
			ClientID: "some-client-id",
			Req:      req,
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 10; 8092) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
		req.RemoteAddr = "192.168.1.1"

		tools.UUIDMock.On("New").Return(uuid.UUID("some-token")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storage.On("Save", mock.Anything, &session).Return(fmt.Errorf("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID:   "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID: "some-client-id",
			Req:      req,
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "failed to save the session: some-error")
	})

	t.Run("GetByToken success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		storage.On("GetByToken", mock.Anything, "some-token").Return(&session, nil).Once()

		res, err := service.GetByToken(ctx, "some-token")
		assert.NoError(t, err)
		assert.EqualValues(t, &session, res)
	})

	t.Run("GetFromReq success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storage.On("GetByToken", mock.Anything, "some-token").Return(&session, nil).Once()

		res, err := service.GetFromReq(req)
		assert.NoError(t, err)
		assert.EqualValues(t, &session, res)
	})

	t.Run("GetFromReq with no cookie", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		// No cookie

		res, err := service.GetFromReq(req)
		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: missing session token")
	})

	t.Run("GetFromReq with the session not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storage.On("GetByToken", mock.Anything, "some-token").Return(nil, nil).Once()

		res, err := service.GetFromReq(req)
		assert.Nil(t, res)
		assert.EqualError(t, err, "bad request: session not found")
	})

	t.Run("GetFromReq with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storage.On("GetByToken", mock.Anything, "some-token").Return(nil, errors.New("some-error")).Once()

		res, err := service.GetFromReq(req)
		assert.Nil(t, res)
		assert.EqualError(t, err, "unhandled error: some-error")
	})

	t.Run("Logout success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storage.On("RemoveByToken", mock.Anything, "some-token").Return(nil).Once()

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
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		// No cookie

		// Do nothing

		err := service.Logout(req, w)
		assert.NoError(t, err)
	})

	t.Run("Logout with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "some-token",
		})

		storage.On("RemoveByToken", mock.Anything, "some-token").Return(errors.New("some-error")).Once()

		err := service.Logout(req, w)
		assert.EqualError(t, err, "failed to remove the token: some-error")
	})

	t.Run("Revoke success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		storage.On("GetByToken", mock.Anything, WebSessionExample.Token()).Return(&WebSessionExample, nil).Once()
		storage.On("RemoveByToken", mock.Anything, WebSessionExample.Token()).Return(nil).Once()

		err := service.Revoke(ctx, &RevokeCmd{
			UserID: WebSessionExample.UserID(),
			Token:  WebSessionExample.Token(),
		})
		assert.NoError(t, err)
	})

	t.Run("Revoke with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		err := service.Revoke(ctx, &RevokeCmd{
			UserID: "some-invalid-id",
			Token:  WebSessionExample.Token(),
		})
		assert.EqualError(t, err, "validation error: UserID: must be a valid UUID v4.")
	})

	t.Run("Revoke with a token not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		storage.On("GetByToken", mock.Anything, WebSessionExample.Token()).Return(nil, nil).Once()

		err := service.Revoke(ctx, &RevokeCmd{
			UserID: WebSessionExample.UserID(),
			Token:  WebSessionExample.Token(),
		})
		assert.NoError(t, err)
	})

	t.Run("Revoke with a token owned by someone else", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		storage.On("GetByToken", mock.Anything, WebSessionExample.Token()).Return(&WebSessionExample, nil).Once()

		err := service.Revoke(ctx, &RevokeCmd{
			UserID: uuid.UUID("29a81212-9e46-4678-a921-ecaf53aa15bc"), // A random user id
			Token:  WebSessionExample.Token(),
		})
		assert.EqualError(t, err, "not found: user ids are not matching")
		assert.ErrorIs(t, err, errs.ErrNotFound)
	})
}
