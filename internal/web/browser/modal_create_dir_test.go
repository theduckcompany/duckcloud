package browser

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
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

func TestCreateDirModal(t *testing.T) {
	t.Run("getCreateDirModel success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/create-dir.tmpl", map[string]any{
			"directory": "/foo/bar",
			"spaceID":   "some-space-id",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		queries.Add("dir", "/foo/bar")
		queries.Add("space", "some-space-id")

		r.URL.RawQuery = queries.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getCreateDirModel with no auth", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrSessionNotFound).Once()

		webSessionsMock.On("Logout", mock.Anything, mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		queries.Add("dir", "/foo/bar")
		queries.Add("space", "some-space-id")

		r.URL.RawQuery = queries.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getCreateDirModel with no dir query", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errors.New("failed to get the dir path from the url query")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		// queries.Add("dir", "/foo/bar")
		queries.Add("space", "some-space-id")

		r.URL.RawQuery = queries.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getCreateDirModel with no dir query", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errors.New("failed to get the space id from the url query")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		queries.Add("dir", "/foo/bar")
		// queries.Add("space", "some-space-id")

		r.URL.RawQuery = queries.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		spaceFSMock.On("Get", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/New Dir"}).Return(nil, errs.ErrNotFound).Once()
		spaceFSMock.On("CreateDir", mock.Anything, &dfs.CreateDirCmd{
			FilePath:  "/foo/New Dir",
			CreatedBy: &users.ExampleAlice,
		}).Return(&dfs.ExampleAliceDir, nil).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
		assert.Equal(t, "refreshFolder", res.Header.Get("HX-Trigger"))
	})

	t.Run("handleCreateDirReq with an authentication error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("handleCreateDirReq with an invalid spaceID", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID(""), fmt.Errorf("some-error")).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errors.New("invalid space id param")).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		// form.Add("name", "New Dir")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with an empty name", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]any{
			"directory": "/foo",
			"spaceID":   uuid.UUID("some-space-id"),
			"error":     "Must not be empty",
		}).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		// form.Add("name", "New Dir")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with a taken name", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		spaceFSMock.On("Get", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/New Dir"}).Return(&dfs.ExampleAliceDir, nil).Once()

		// Render
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]any{
			"directory": "/foo",
			"spaceID":   uuid.UUID("some-space-id"),
			"error":     "Already exists",
		}).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with a CreateDir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newCreateDirModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		err := fmt.Errorf("some-error")
		spaceFSMock.On("Get", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/New Dir"}).Return(nil, errs.ErrNotFound).Once()
		spaceFSMock.On("CreateDir", mock.Anything, &dfs.CreateDirCmd{
			FilePath:  "/foo/New Dir",
			CreatedBy: &users.ExampleAlice,
		}).Return(nil, err).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to create the directory: %w", err)).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})
}
