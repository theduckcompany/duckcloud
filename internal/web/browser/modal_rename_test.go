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
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
)

func Test_RenameModalHandler(t *testing.T) {
	t.Run("getRenameModal success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newRenameModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar.jpg")).Return(&dfs.ExampleAliceFile, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &browser.RenameTemplate{
			Error:               nil,
			Target:              dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar.jpg"),
			FieldValue:          "bar.jpg",
			FieldValueSelection: 3, // name == bar.pdf / we want the selection at |bar|.pdf
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/rename", nil)
		r.URL.RawQuery = url.Values{
			"path":    []string{"/foo/bar.jpg"},
			"value":   []string{"bar.jpg"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getRenameModal with an unauthenticated user", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newRenameModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/rename", nil)
		r.URL.RawQuery = url.Values{
			"path":    []string{"/foo/bar.jpg"},
			"value":   []string{"bar.jpg"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("getRenameModal without the path param", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newRenameModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/rename", nil)
		r.URL.RawQuery = url.Values{
			// "path":    []string{"/foo/bar.jpg"},
			"value":   []string{"bar.jpg"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("getRenameModal with an invalid spaceID", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newRenameModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID(""), fmt.Errorf("some-error")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/rename", nil)
		r.URL.RawQuery = url.Values{
			"path":    []string{"/foo/bar.jpg"},
			"value":   []string{"bar.jpg"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("handleRenameReq success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newRenameModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar.jpg")).Return(&dfs.ExampleAliceFile, nil).Once()

		fsMock.On("Rename", mock.Anything, &dfs.ExampleAliceFile, "new-name.jpg").Return(&dfs.ExampleAliceFile, nil).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("path", "/foo/bar.jpg")
		form.Add("name", "new-name.jpg")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/rename", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "refreshPage", res.Header.Get("HX-Trigger"))
	})

	t.Run("handleRenameReq with a rename error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newRenameModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar.jpg")).Return(&dfs.ExampleAliceFile, nil).Once()

		fsMock.On("Rename", mock.Anything, &dfs.ExampleAliceFile, "new-name").Return(nil, errs.Validation(errors.New("some-error"))).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, &browser.RenameTemplate{
			Error:               ptr.To("validation: some-error"),
			Target:              dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar.jpg"),
			FieldValue:          "new-name",
			FieldValueSelection: 8, // name == new-name / we want the selection at |new-name|.pdf
		}).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("path", "/foo/bar.jpg")
		form.Add("name", "new-name")
		form.Add("spaceID", "some-space-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/rename", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
