package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
)

func TestHTTHandler_createUser(t *testing.T) {
	tools := tools.NewMock(t)
	service := NewMockService(t)

	handler := NewHTTPHandler(tools, service)
	r := chi.NewRouter()
	handler.Register(r, router.InitMiddlewares(tools))
	server := httptest.NewServer(r)
	defer server.Close()

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

	e := httpexpect.Default(t, server.URL)

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
}
