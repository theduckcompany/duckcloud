package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

func Test_Settings(t *testing.T) {
	t.Run("getSecurityPage success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		webSessionsMock.On("GetAllForUser", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]websessions.Session{websessions.AliceWebSessionExample}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, users.ExampleAlice.ID(), &storage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{davsessions.ExampleAliceSession}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]spaces.Space{spaces.ExampleAlicePersonalSpace}, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "settings/security/content.tmpl", map[string]interface{}{
			"isAdmin":        users.ExampleAlice.IsAdmin(),
			"currentSession": &websessions.AliceWebSessionExample,
			"webSessions":    []websessions.Session{websessions.AliceWebSessionExample},
			"devices":        []davsessions.DavSession{davsessions.ExampleAliceSession},
			"spaces": map[uuid.UUID]spaces.Space{
				spaces.ExampleAlicePersonalSpace.ID(): spaces.ExampleAlicePersonalSpace,
			},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/security", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getSecurityPage redirect to login if not authenticated", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/security", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("createDavSession success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "space-id-1").Return(uuid.UUID("space-id-1"), nil).Once()

		davSessionsMock.On("Create", mock.Anything, &davsessions.CreateCmd{
			UserID:   users.ExampleAlice.ID(),
			Name:     "some dav-session name",
			Username: users.ExampleAlice.Username(),
			SpaceID:  uuid.UUID("space-id-1"),
		}).Return(&davsessions.ExampleAliceSession2, "some-session-secret", nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusCreated, "settings/security/webdav-result.tmpl", map[string]interface{}{
			"newSession": &davsessions.ExampleAliceSession2,
			"secret":     "some-session-secret",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/webdav", strings.NewReader(url.Values{
			"space": []string{"space-id-1"},
			"name":  []string{"some dav-session name"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("deleteWebSession success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-token").Return(uuid.UUID("some-token"), nil).Once()

		webSessionsMock.On("Delete", mock.Anything, &websessions.DeleteCmd{
			UserID: users.ExampleAlice.ID(),
			Token:  secret.NewText("some-token"),
		}).Return(nil).Once()

		webSessionsMock.On("GetAllForUser", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]websessions.Session{websessions.AliceWebSessionExample}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, users.ExampleAlice.ID(), &storage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{davsessions.ExampleAliceSession}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]spaces.Space{spaces.ExampleAlicePersonalSpace}, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "settings/security/content.tmpl", map[string]interface{}{
			"isAdmin":        users.ExampleAlice.IsAdmin(),
			"currentSession": &websessions.AliceWebSessionExample,
			"webSessions":    []websessions.Session{websessions.AliceWebSessionExample},
			"devices":        []davsessions.DavSession{davsessions.ExampleAliceSession},
			"spaces": map[uuid.UUID]spaces.Space{
				spaces.ExampleAlicePersonalSpace.ID(): spaces.ExampleAlicePersonalSpace,
			},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/browsers/some-token/delete", nil)

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("deleteDavSession success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-session-id").Return(uuid.UUID("some-session-id"), nil).Once()

		davSessionsMock.On("Delete", mock.Anything, &davsessions.DeleteCmd{
			UserID:    users.ExampleAlice.ID(),
			SessionID: uuid.UUID("some-session-id"),
		}).Return(nil).Once()

		webSessionsMock.On("GetAllForUser", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]websessions.Session{websessions.AliceWebSessionExample}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, users.ExampleAlice.ID(), &storage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{davsessions.ExampleAliceSession}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]spaces.Space{spaces.ExampleAlicePersonalSpace}, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "settings/security/content.tmpl", map[string]interface{}{
			"isAdmin":        users.ExampleAlice.IsAdmin(),
			"currentSession": &websessions.AliceWebSessionExample,
			"webSessions":    []websessions.Session{websessions.AliceWebSessionExample},
			"devices":        []davsessions.DavSession{davsessions.ExampleAliceSession},
			"spaces": map[uuid.UUID]spaces.Space{
				spaces.ExampleAlicePersonalSpace.ID(): spaces.ExampleAlicePersonalSpace,
			},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/webdav/some-session-id/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getUsers success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, &storage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "settings/users/content.tmpl", map[string]interface{}{
			"isAdmin": users.ExampleAlice.IsAdmin(),
			"current": &users.ExampleAlice,
			"users":   []users.User{users.ExampleAlice, users.ExampleBob},
			"error":   nil,
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
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-user-id").Return(uuid.UUID("some-user-id"), nil).Once()

		usersMock.On("AddToDeletion", mock.Anything, uuid.UUID("some-user-id")).Return(nil).Once()

		usersMock.On("GetAll", mock.Anything, &storage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "settings/users/content.tmpl", map[string]interface{}{
			"isAdmin": users.ExampleAlice.IsAdmin(),
			"current": &users.ExampleAlice,
			"users":   []users.User{users.ExampleAlice, users.ExampleBob},
			"error":   nil,
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
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newSettingsHandler(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("Create", mock.Anything, &users.CreateCmd{
			Username: "some-username",
			Password: secret.NewText("my-little-secret"),
			IsAdmin:  true,
		}).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, &storage.PaginateCmd{
			StartAfter: map[string]string{"username": ""},
			Limit:      20,
		}).Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "settings/users/content.tmpl", map[string]interface{}{
			"isAdmin": users.ExampleAlice.IsAdmin(),
			"current": &users.ExampleAlice,
			"users":   []users.User{users.ExampleAlice, users.ExampleBob},
			"error":   nil,
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
