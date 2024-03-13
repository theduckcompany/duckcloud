package settings

import (
	"errors"
	"fmt"
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
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/settings/security"
)

func Test_SecurityPage(t *testing.T) {
	t.Parallel()

	t.Run("getSecurityPage success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).CreatedBy(user).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		davSession := davsessions.NewFakeSession(t).CreatedBy(user).WithSpace(space).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		webSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]websessions.Session{*webSession}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), &sqlstorage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{*davSession}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]spaces.Space{*space}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &security.ContentTemplate{
			IsAdmin:        user.IsAdmin(),
			CurrentSession: webSession,
			WebSessions:    []websessions.Session{*webSession},
			Devices:        []davsessions.DavSession{*davSession},
			Spaces:         map[uuid.UUID]spaces.Space{space.ID(): *space},
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/security", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getSecurityPage redirect to login if not authenticated", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/security", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("createDavSession success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).CreatedBy(user).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		newSessionSecret := "some-secret"
		newDavSession := davsessions.NewFakeSession(t).
			CreatedBy(user).
			WithSpace(space).
			WithPassword(newSessionSecret).
			WithName("some dav-session name").
			Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", string(space.ID())).Return(space.ID(), nil).Once()
		davSessionsMock.On("Create", mock.Anything, &davsessions.CreateCmd{
			UserID:   user.ID(),
			Name:     "some dav-session name",
			Username: user.Username(),
			SpaceID:  space.ID(),
		}).Return(newDavSession, newSessionSecret, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusCreated, &security.WebdavResultTemplate{
			NewSession: newDavSession,
			Secret:     newSessionSecret,
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/webdav", strings.NewReader(url.Values{
			"space": []string{string(space.ID())},
			"name":  []string{"some dav-session name"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Assert
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteWebSession success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		space := spaces.NewFakeSpace(t).CreatedBy(user).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		davSession := davsessions.NewFakeSession(t).CreatedBy(user).WithSpace(space).Build()
		session2Token := "some-token"

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", session2Token).Return(uuid.UUID(session2Token), nil).Once()
		webSessionsMock.On("Delete", mock.Anything, &websessions.DeleteCmd{
			UserID: user.ID(),
			Token:  secret.NewText(session2Token),
		}).Return(nil).Once()

		webSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]websessions.Session{*webSession}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), &sqlstorage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{*davSession}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]spaces.Space{*space}, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &security.ContentTemplate{
			IsAdmin:        user.IsAdmin(),
			CurrentSession: webSession,
			WebSessions:    []websessions.Session{*webSession},
			Devices:        []davsessions.DavSession{*davSession},
			Spaces:         map[uuid.UUID]spaces.Space{space.ID(): *space},
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/browsers/"+session2Token+"/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Assert
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteDavSession success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Authentication
		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		space := spaces.NewFakeSpace(t).CreatedBy(user).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		davSession := davsessions.NewFakeSession(t).CreatedBy(user).WithSpace(space).Build()
		davSessionToken := "some-dav-session-token"

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", davSessionToken).Return(uuid.UUID(davSessionToken), nil).Once()

		davSessionsMock.On("Delete", mock.Anything, &davsessions.DeleteCmd{
			UserID:    user.ID(),
			SessionID: uuid.UUID(davSessionToken),
		}).Return(nil).Once()

		webSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]websessions.Session{*webSession}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), &sqlstorage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{*davSession}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]spaces.Space{*space}, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &security.ContentTemplate{
			IsAdmin:        user.IsAdmin(),
			CurrentSession: webSession,
			WebSessions:    []websessions.Session{*webSession},
			Devices:        []davsessions.DavSession{*davSession},
			Spaces:         map[uuid.UUID]spaces.Space{space.ID(): *space},
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/webdav/"+davSessionToken+"/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("updatePassword success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("old-password")).
			Return(user, nil).Once()

		usersMock.On("UpdateUserPassword", mock.Anything, &users.UpdatePasswordCmd{
			UserID:      user.ID(),
			NewPassword: secret.NewText("new-password"),
		}).Return(nil).Once()

		// Print the security page
		webSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]websessions.Session{*webSession}, nil).Once()
		davSessionsMock.On("GetAllForUser", mock.Anything, user.ID(), &sqlstorage.PaginateCmd{Limit: 20}).Return([]davsessions.DavSession{}, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, user.ID(), (*sqlstorage.PaginateCmd)(nil)).Return([]spaces.Space{}, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &security.ContentTemplate{
			IsAdmin:        user.IsAdmin(),
			CurrentSession: webSession,
			WebSessions:    []websessions.Session{*webSession},
			Devices:        []davsessions.DavSession{},
			Spaces:         map[uuid.UUID]spaces.Space{},
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/password", strings.NewReader(url.Values{
			"current": []string{"old-password"},
			"new":     []string{"new-password"},
			"confirm": []string{"new-password"},
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

	t.Run("updatePassword with an invalid current password", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("old-password")).
			Return(nil, users.ErrInvalidPassword).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, &security.PasswordFormTemplate{
			Error: "invalid current password",
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/password", strings.NewReader(url.Values{
			"current": []string{"old-password"},
			"new":     []string{"new-password"},
			"confirm": []string{"new-password"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Assert
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("updatePassword with an invalid confirmation password", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("old-password")).
			Return(user, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, &security.PasswordFormTemplate{
			Error: "the new password and the confirmation are different",
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/password", strings.NewReader(url.Values{
			"current": []string{"old-password"},
			"new":     []string{"new-password"},
			"confirm": []string{"different"}, // <<<<<<<<<< Should be equal "new-password"
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

	t.Run("updatePassword with an authentication error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("old-password")).
			Return(nil, errs.Internal(errors.New("some-error"))).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to Authenticate the user: %w", errs.Internal(errors.New("some-error")))).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/password", strings.NewReader(url.Values{
			"current": []string{"old-password"},
			"new":     []string{"new-password"},
			"confirm": []string{"new-password"},
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

	t.Run("updatePassword with an update password error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("old-password")).
			Return(user, nil).Once()

		usersMock.On("UpdateUserPassword", mock.Anything, &users.UpdatePasswordCmd{
			UserID:      user.ID(),
			NewPassword: secret.NewText("new-password"),
		}).Return(errs.Validation(fmt.Errorf("some-error"))).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, &security.PasswordFormTemplate{
			Error: "validation: some-error",
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/security/password", strings.NewReader(url.Values{
			"current": []string{"old-password"},
			"new":     []string{"new-password"},
			"confirm": []string{"new-password"},
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

	t.Run("getPasswordForm success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &security.PasswordFormTemplate{
			Error: "",
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/security/password", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getPasswordForm redirect to login if not authenticated", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSecurityPage(tools, htmlMock, webSessionsMock, davSessionsMock, spacesMock, usersMock, auth)

		// Data

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/security/password", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})
}
