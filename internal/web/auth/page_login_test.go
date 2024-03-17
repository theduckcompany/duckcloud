package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

func Test_LoginPage(t *testing.T) {
	t.Run("Login without any session open", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything).Return(nil, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.LoginPageTmpl{})

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/login", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Login while already being authenticated redirect to the home", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		webSession := websessions.NewFakeSession(t).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything).Return(webSession, nil).Once()
		tools.UUIDMock.On("Parse", "").Return(uuid.UUID(""), errors.New("invalid")).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/login", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/", res.Header.Get("Location"))
	})

	t.Run("Login step from the oauth2 dance redirect to the consent page", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		webSession := websessions.NewFakeSession(t).Build()
		oauthClient := oauthclients.NewFakeClient(t).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything).Return(webSession, nil).Once()
		tools.UUIDMock.On("Parse", oauthClient.GetID()).Return(uuid.UUID(oauthClient.GetID()), nil).Once()
		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID(oauthClient.GetID())).Return(oauthClient, nil).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/login", nil)
		vals := url.Values{}
		vals.Add("client_id", oauthClient.GetID())
		vals.Add("another_field", "some-content")
		r.URL.RawQuery = vals.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/consent?"+vals.Encode(), res.Header.Get("Location"))
	})

	t.Run("Login step from the oauth2 dance with skip-validation redirect to the /authorize endpoint", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		webSession := websessions.NewFakeSession(t).Build()
		oauthClient := oauthclients.NewFakeClient(t).SkipValidation().Build() // NOTE: SkipValidation

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything).Return(webSession, nil).Once()
		tools.UUIDMock.On("Parse", oauthClient.GetID()).Return(uuid.UUID(oauthClient.GetID()), nil).Once()
		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID(oauthClient.GetID())).Return(oauthClient, nil).Once()

		// run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/login", nil)
		vals := url.Values{}
		vals.Add("client_id", oauthClient.GetID())
		vals.Add("another_field", "some-content")
		r.URL.RawQuery = vals.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/auth/authorize", res.Header.Get("Location"))
	})

	t.Run("ApplyLogin success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		userPassword := gofakeit.Password(true, true, true, false, false, 8)
		user := users.NewFakeUser(t).WithPassword(userPassword).Build()
		webSession := websessions.NewFakeSession(t).
			CreatedBy(user).
			WithDevice("firefox 4.4.4.4").
			WithIP(httptest.DefaultRemoteAddr).
			Build()

		// Mocks
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText(userPassword)).
			Return(user, nil).Once()
		webSessionsMock.On("Create", mock.Anything, &websessions.CreateCmd{
			UserID:     user.ID(),
			UserAgent:  "firefox 4.4.4.4",
			RemoteAddr: httptest.DefaultRemoteAddr,
		}).Return(webSession, nil).Once()
		tools.UUIDMock.On("Parse", "").Return(uuid.UUID(""), errors.New("invalid")).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
			"username": []string{user.Username()},
			"password": []string{userPassword},
		}.Encode()))
		r.RemoteAddr = httptest.DefaultRemoteAddr
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", "firefox 4.4.4.4")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/", res.Header.Get("Location"))
		assert.Len(t, res.Cookies(), 1)
		assert.Equal(t, "session_token", res.Cookies()[0].Name)
		assert.Empty(t, res.Cookies()[0].Expires)
	})

	t.Run("ApplyLogin with an invalid username", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data

		// Mocks
		usersMock.On("Authenticate", mock.Anything, "invalid-username", secret.NewText("some-password")).
			Return(nil, users.ErrInvalidUsername).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusBadRequest, &auth.LoginPageTmpl{
			UsernameContent: "invalid-username",
			UsernameError:   "User doesn't exists",
			PasswordError:   "",
		})

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
			"username": []string{"invalid-username"},
			"password": []string{"some-password"},
		}.Encode()))
		r.RemoteAddr = httptest.DefaultRemoteAddr
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", "firefox 4.4.4.4")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("ApplyLogin with an invalid password", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()

		// Mocks
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("some-invalid-password")).
			Return(nil, users.ErrInvalidPassword).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusBadRequest, &auth.LoginPageTmpl{
			UsernameContent: user.Username(),
			UsernameError:   "",
			PasswordError:   "Invalid password",
		})

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
			"username": []string{user.Username()},
			"password": []string{"some-invalid-password"},
		}.Encode()))
		r.RemoteAddr = httptest.DefaultRemoteAddr
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", "firefox 4.4.4.4")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("ApplyLogin with an authentication error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		user := users.NewFakeUser(t).Build()

		// Mocks
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText("some-invalid-password")).
			Return(nil, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errs.ErrInternal)

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
			"username": []string{user.Username()},
			"password": []string{"some-invalid-password"},
		}.Encode()))
		r.RemoteAddr = httptest.DefaultRemoteAddr
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", "firefox 4.4.4.4")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("ApplyLogin during a oauth2 dance", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		userPassword := gofakeit.Password(true, true, true, false, false, 8)
		user := users.NewFakeUser(t).WithPassword(userPassword).Build()
		oauthClient := oauthclients.NewFakeClient(t).CreatedBy(user).SkipValidation().Build()
		webSession := websessions.NewFakeSession(t).
			CreatedBy(user).
			WithDevice("firefox 4.4.4.4").
			WithIP(httptest.DefaultRemoteAddr).
			Build()

		// Mocks
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText(userPassword)).
			Return(user, nil).Once()
		webSessionsMock.On("Create", mock.Anything, &websessions.CreateCmd{
			UserID:     user.ID(),
			UserAgent:  "firefox 4.4.4.4",
			RemoteAddr: httptest.DefaultRemoteAddr,
		}).Return(webSession, nil).Once()
		tools.UUIDMock.On("Parse", oauthClient.GetID()).Return(uuid.UUID(oauthClient.GetID()), nil).Once()
		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID(oauthClient.GetID())).
			Return(oauthClient, nil).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
			"username": []string{user.Username()},
			"password": []string{userPassword},
		}.Encode()))
		r.RemoteAddr = httptest.DefaultRemoteAddr
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", "firefox 4.4.4.4")

		vals := url.Values{}
		vals.Add("client_id", oauthClient.GetID())
		vals.Add("another_field", "some-content")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/auth/authorize", res.Header.Get("Location"))
		assert.Len(t, res.Cookies(), 1)
		assert.Equal(t, "session_token", res.Cookies()[0].Name)
	})

	t.Run("ApplyLogin during a oauth2 dance with a clients.GetByID error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		handler := NewLoginPage(htmlMock, webSessionsMock, usersMock, oauthclientsMock, tools)

		// Data
		userPassword := gofakeit.Password(true, true, true, false, false, 8)
		user := users.NewFakeUser(t).WithPassword(userPassword).Build()
		oauthClient := oauthclients.NewFakeClient(t).CreatedBy(user).SkipValidation().Build()
		webSession := websessions.NewFakeSession(t).
			CreatedBy(user).
			WithDevice("firefox 4.4.4.4").
			WithIP(httptest.DefaultRemoteAddr).
			Build()

		// Mocks
		usersMock.On("Authenticate", mock.Anything, user.Username(), secret.NewText(userPassword)).
			Return(user, nil).Once()
		webSessionsMock.On("Create", mock.Anything, &websessions.CreateCmd{
			UserID:     user.ID(),
			UserAgent:  "firefox 4.4.4.4",
			RemoteAddr: httptest.DefaultRemoteAddr,
		}).Return(webSession, nil).Once()
		tools.UUIDMock.On("Parse", oauthClient.GetID()).Return(uuid.UUID(oauthClient.GetID()), nil).Once()
		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID(oauthClient.GetID())).
			Return(nil, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusBadRequest, &auth.ErrorPageTmpl{
			ErrorMsg:  "Oauth client not found",
			RequestID: "????",
		}).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
			"username": []string{user.Username()},
			"password": []string{userPassword},
		}.Encode()))
		r.RemoteAddr = httptest.DefaultRemoteAddr
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", "firefox 4.4.4.4")

		vals := url.Values{}
		vals.Add("client_id", oauthClient.GetID())
		vals.Add("another_field", "some-content")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
