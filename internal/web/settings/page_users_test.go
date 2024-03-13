package settings

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	userstmpl "github.com/theduckcompany/duckcloud/internal/web/html/templates/settings/users"
)

func Test_UsersPage(t *testing.T) {
	t.Parallel()

	t.Run("getUsers success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewUsersPage(tools, htmlMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		user2 := users.NewFakeUser(t).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("GetAll", mock.Anything, &sqlstorage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{*user, *user2}, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &userstmpl.ContentTemplate{
			IsAdmin: user.IsAdmin(),
			Current: user,
			Users:   []users.User{*user, *user2},
			Error:   nil,
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/users", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteUser success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewUsersPage(tools, htmlMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		user2 := users.NewFakeUser(t).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		someUserID := "some-user-id"

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", someUserID).Return(uuid.UUID(someUserID), nil).Once()

		usersMock.On("AddToDeletion", mock.Anything, uuid.UUID(someUserID)).Return(nil).Once()

		usersMock.On("GetAll", mock.Anything, &sqlstorage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{*user, *user2}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &userstmpl.ContentTemplate{
			IsAdmin: user.IsAdmin(),
			Current: user,
			Users:   []users.User{*user, *user2},
			Error:   nil,
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/users/"+someUserID+"/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("createUser success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewUsersPage(tools, htmlMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		newUserPassword := "some-password"
		newUser := users.NewFakeUser(t).
			WithUsername("Alice").
			WithPassword(newUserPassword).
			WithAdminRole().
			Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("Create", mock.Anything, &users.CreateCmd{
			CreatedBy: user,
			Username:  "Alice",
			Password:  secret.NewText(newUserPassword),
			IsAdmin:   true,
		}).Return(newUser, nil).Once()
		usersMock.On("GetAll", mock.Anything, &sqlstorage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{*user, *newUser}, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &userstmpl.ContentTemplate{
			IsAdmin: user.IsAdmin(),
			Current: user,
			Users:   []users.User{*user, *newUser},
			Error:   nil,
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/users", strings.NewReader(url.Values{
			"username": []string{"Alice"},
			"password": []string{newUserPassword},
			"role":     []string{"admin"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
