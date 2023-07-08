package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
)

func TestHTTHandler(t *testing.T) {
	t.Run("createUser success", func(t *testing.T) {
		_, service, e := setupRouter(t)

		now := time.Now()

		service.On("Create", mock.Anything, &CreateCmd{
			Username: "some-username",
			Email:    "some-email",
			Password: "some-password",
		}).Return(&User{
			Username:  "some-username",
			Email:     "some-email",
			CreatedAt: now,
			password:  "some-password",
		}, nil).Once()

		obj := e.POST("/users").
			WithHeader("Content-Type", "application/json").
			WithBytes([]byte(`{
      "username": "some-username",
      "email": "some-email",
      "password": "some-password"
    }`)).
			Expect().Status(http.StatusCreated).
			JSON().Schema(`{
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "username": { "type": "string" },
        "email": { "type": "string" },
        "createdAt": { "type": "string" }
      }
    }`).Object()

		obj.HasValue("username", "some-username")
		obj.HasValue("email", "some-email")
		obj.HasValue("createdAt", now.Format(time.RFC3339Nano))
	})

	t.Run("getMyUser success", func(t *testing.T) {
		tools, service, e := setupRouter(t)

		now := time.Now()

		tools.JWTMock.On("FetchAccessToken", mock.Anything).Return(&jwt.AccessToken{
			ClientID: uuid.UUID("some-client-id"),
			UserID:   uuid.UUID("some-user-id"),
			Raw:      "some-raw-token",
		}, nil).Once()

		service.On("GetByID", mock.Anything, uuid.UUID("some-user-id")).Return(&User{
			ID:        uuid.UUID("some-user-id"),
			Username:  "some-username",
			Email:     "some-email",
			CreatedAt: now,
			password:  "some-password",
		}, nil).Once()

		obj := e.GET("/users/me").
			Expect().Status(http.StatusOK).
			JSON().Schema(`{
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "username": { "type": "string" },
        "email": { "type": "string" },
        "createdAt": { "type": "string" }
      }
    }`).Object()

		obj.HasValue("username", "some-username")
		obj.HasValue("email", "some-email")
		obj.HasValue("createdAt", now.Format(time.RFC3339Nano))
	})
}

func setupRouter(t *testing.T) (*tools.Mock, *MockService, *httpexpect.Expect) {
	tools := tools.NewMock(t)
	service := NewMockService(t)

	handler := NewHTTPHandler(tools, service)
	r := chi.NewRouter()
	handler.Register(r, router.InitMiddlewares(tools))
	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	e := httpexpect.Default(t, server.URL)

	return tools, service, e
}
