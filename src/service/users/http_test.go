package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi/v5"
	"github.com/myminicloud/myminicloud/src/service/oauth2"
	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/router"
	"github.com/myminicloud/myminicloud/src/tools/uuid"
	"github.com/stretchr/testify/mock"
)

func TestHTTHandler(t *testing.T) {
	t.Run("createUser success", func(t *testing.T) {
		service, _, e := setupRouter(t)

		now := time.Now()

		service.On("Create", mock.Anything, &CreateCmd{
			Username: "some-username",
			Email:    "some-email",
			Password: "some-password",
		}).Return(&User{
			username:  "some-username",
			email:     "some-email",
			createdAt: now,
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
		service, oauth2Svc, e := setupRouter(t)

		now := time.Now()

		oauth2Svc.On("GetFromReq", mock.Anything).Return(&oauth2.Token{
			UserID: uuid.UUID("some-user-id"),
		}, nil).Once()

		service.On("GetByID", mock.Anything, uuid.UUID("some-user-id")).Return(&User{
			id:        uuid.UUID("some-user-id"),
			username:  "some-username",
			email:     "some-email",
			createdAt: now,
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

func setupRouter(t *testing.T) (*MockService, *oauth2.MockService, *httpexpect.Expect) {
	tools := tools.NewMock(t)
	oauth2Svc := oauth2.NewMockService(t)
	service := NewMockService(t)

	handler := NewHTTPHandler(tools, service, oauth2Svc)
	r := chi.NewRouter()
	handler.Register(r, router.InitMiddlewares(tools))
	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	e := httpexpect.Default(t, server.URL)

	return service, oauth2Svc, e
}
