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
	t.Run("getUsers success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewUsersPage(tools, htmlMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, &storage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &userstmpl.ContentTemplate{
			IsAdmin: users.ExampleAlice.IsAdmin(),
			Current: &users.ExampleAlice,
			Users:   []users.User{users.ExampleAlice, users.ExampleBob},
			Error:   nil,
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/users", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewUsersPage(tools, htmlMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-user-id").Return(uuid.UUID("some-user-id"), nil).Once()

		usersMock.On("AddToDeletion", mock.Anything, uuid.UUID("some-user-id")).Return(nil).Once()

		usersMock.On("GetAll", mock.Anything, &storage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &userstmpl.ContentTemplate{
			IsAdmin: users.ExampleAlice.IsAdmin(),
			Current: &users.ExampleAlice,
			Users:   []users.User{users.ExampleAlice, users.ExampleBob},
			Error:   nil,
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/users/some-user-id/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("createUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewUsersPage(tools, htmlMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("Create", mock.Anything, &users.CreateCmd{
			CreatedBy: &users.ExampleAlice,
			Username:  "some-username",
			Password:  secret.NewText("my-little-secret"),
			IsAdmin:   true,
		}).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, &storage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &userstmpl.ContentTemplate{
			IsAdmin: users.ExampleAlice.IsAdmin(),
			Current: &users.ExampleAlice,
			Users:   []users.User{users.ExampleAlice, users.ExampleBob},
			Error:   nil,
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/users", strings.NewReader(url.Values{
			"username": []string{"some-username"},
			"password": []string{"my-little-secret"},
			"role":     []string{"admin"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
