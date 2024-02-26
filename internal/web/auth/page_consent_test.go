package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

func Test_ConsentPage(t *testing.T) {
	t.Run("Print the consent page success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()

		tools.UUIDMock.On("Parse", "some-client-id").Return(uuid.UUID("some-client-id"), nil).Once()

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID("some-client-id")).
			Return(&oauthclients.ExampleAliceClient, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &auth.ConsentPageTmpl{
			Username:   users.ExampleAlice.Username(),
			Redirect:   "/consent?client_id=some-client-id&scope=scope-a%2Cscope-b",
			ClientName: oauthclients.ExampleAliceClient.Name(),
			Scopes:     []string{"scope-a", "scope-b"},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Print the consent page with no sessions", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrSessionNotFound).Once()

		webSessionsMock.On("Logout", mock.Anything, mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Print the consent page with an invalid client_id format", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()
		tools.UUIDMock.On("Parse", "some-client-id").Return(uuid.UUID(""), errors.New("invalid")).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusBadRequest, &auth.ErrorPageTmpl{
			ErrorMsg:  "invalid client_id",
			RequestID: "????",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Print the consent page with a client not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()

		tools.UUIDMock.On("Parse", "some-client-id").Return(uuid.UUID("some-client-id"), nil).Once()

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID("some-client-id")).
			Return(nil, errs.ErrNotFound).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusBadRequest, &auth.ErrorPageTmpl{
			ErrorMsg:  "invalid client_id",
			RequestID: "????",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Print the consent page with a client.GetByID error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()

		tools.UUIDMock.On("Parse", "some-client-id").Return(uuid.UUID("some-client-id"), nil).Once()

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID("some-client-id")).
			Return(nil, errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errs.ErrInternal).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Validate the consent page success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()

		tools.UUIDMock.On("Parse", "some-client-id").Return(uuid.UUID("some-client-id"), nil).Once()

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID("some-client-id")).
			Return(&oauthclients.ExampleAliceClient, nil).Once()

		oauthConsentMock.On("Create", mock.Anything, &oauthconsents.CreateCmd{
			UserID:       users.ExampleAlice.ID(),
			SessionToken: websessions.AliceWebSessionExample.Token().Raw(),
			ClientID:     oauthconsents.ExampleAliceConsent.ClientID(),
			Scopes:       []string{"scope-a", "scope-b"},
		}).Return(&oauthconsents.ExampleAliceConsent, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/auth/authorize?client_id=some-client-id&consent_id=01ce56b3-5ab9-4265-b1d2-e0347dcd4158&scope=scope-a%2Cscope-b", res.Header.Get("Location"))
	})

	t.Run("Validate the consent page with a consent creation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		oauthclientsMock := oauthclients.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		authenticator := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := newConsentPage(htmlMock, authenticator, oauthclientsMock, oauthConsentMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()

		tools.UUIDMock.On("Parse", "some-client-id").Return(uuid.UUID("some-client-id"), nil).Once()

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		oauthclientsMock.On("GetByID", mock.Anything, uuid.UUID("some-client-id")).
			Return(&oauthclients.ExampleAliceClient, nil).Once()

		oauthConsentMock.On("Create", mock.Anything, &oauthconsents.CreateCmd{
			UserID:       users.ExampleAlice.ID(),
			SessionToken: websessions.AliceWebSessionExample.Token().Raw(),
			ClientID:     oauthconsents.ExampleAliceConsent.ClientID(),
			Scopes:       []string{"scope-a", "scope-b"},
		}).Return(nil, errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errs.ErrInternal)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/consent", nil)
		vals := url.Values{}
		vals.Add("client_id", "some-client-id")
		vals.Add("scope", "scope-a,scope-b")
		r.URL.RawQuery = vals.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
