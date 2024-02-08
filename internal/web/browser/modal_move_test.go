package browser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
)

func Test_MoveModalHandler(t *testing.T) {
	t.Run("getMoveModal success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": ""}, Limit: PageSize}).
			Return([]dfs.INode{dfs.ExampleAliceFile2}, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg")).
			Return(&dfs.ExampleAliceFile, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &browser.MoveTemplate{
			SrcPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			SrcInode:      &dfs.ExampleAliceFile,
			DstPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			FolderContent: map[dfs.PathCmd]dfs.INode{*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.txt"): dfs.ExampleAliceFile2},
			PageSize:      PageSize,
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getMoveModal with more content success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": "some-file-name.jpg"}, Limit: PageSize}).
			Return([]dfs.INode{dfs.ExampleAliceFile2}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &browser.MoveRowsTemplate{
			SrcPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			DstPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			FolderContent: map[dfs.PathCmd]dfs.INode{*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.txt"): dfs.ExampleAliceFile2},
			PageSize:      PageSize,
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
			"last":    []string{"some-file-name.jpg"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getMoveModal with an authentication error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
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

	t.Run("getMoveModal with an invalid spaceID", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID(""), fmt.Errorf("some-error")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("getMoveModal with an space not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(nil, errs.ErrNotFound).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to get the space: %w", errs.ErrNotFound))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getMoveModa/getMoveReq with a missing dstPath arg", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			// "dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("getMoveModa/getMoveReql with a missing srcPath arg", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			// "srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("renderMoveModal", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": ""}, Limit: PageSize}).
			Return([]dfs.INode{dfs.ExampleAliceFile2}, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg")).
			Return(&dfs.ExampleAliceFile, nil).Once()

		htmlMock.On("WriteHTMLTemplate", w, r, http.StatusOK, &browser.MoveTemplate{
			SrcPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			SrcInode:      &dfs.ExampleAliceFile,
			DstPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			FolderContent: map[dfs.PathCmd]dfs.INode{*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.txt"): dfs.ExampleAliceFile2},
			PageSize:      PageSize,
		})

		handler.renderMoveModal(w, r, &moveModalCmd{
			Src: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
		})

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("renderMoveModal with a ListDir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": ""}, Limit: PageSize}).
			Return(nil, errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", w, r, fmt.Errorf("failed to list dir for elem /bar: %w", errs.ErrInternal))

		handler.renderMoveModal(w, r, &moveModalCmd{
			Src: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
		})

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("renderMoveModal with a Get error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": ""}, Limit: PageSize}).
			Return([]dfs.INode{dfs.ExampleAliceFile2}, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg")).
			Return(nil, errs.ErrNotFound).Once()

		htmlMock.On("WriteHTMLErrorPage", w, r, fmt.Errorf("failed to get the source file /foo/file.jpg: %w", errs.ErrNotFound))

		handler.renderMoveModal(w, r, &moveModalCmd{
			Src: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
		})

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("renderMoreContent with a ListDir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/move", nil)

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": "some-file-name.jpg"}, Limit: PageSize}).
			Return(nil, errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", w, r, fmt.Errorf("failed to ListDir: %w", errs.ErrInternal))

		handler.renderMoreContent(w, r, &moveModalCmd{
			Src: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst: dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
		}, "some-file-name.jpg")

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("handleMoveReq success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Move", mock.Anything, &dfs.MoveCmd{
			Src:     dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst:     dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.jpg"),
			MovedBy: &users.ExampleAlice,
		}).Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, "none", res.Header.Get("HX-Reswap"))
		assert.Equal(t, "refreshPage", res.Header.Get("HX-Trigger"))
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("handleMoveReq with a move error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Move", mock.Anything, &dfs.MoveCmd{
			Src:     dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst:     dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.jpg"),
			MovedBy: &users.ExampleAlice,
		}).Return(errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to move the file: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleMoveReq with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newMoveModalHandler(auth, spacesMock, htmlMock, tools.UUID(), fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space
		tools.UUIDMock.On("Parse", "some-space-id").Return(uuid.UUID("some-space-id"), nil).Once()
		spacesMock.On("GetByID", mock.Anything, uuid.UUID("some-space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Move", mock.Anything, &dfs.MoveCmd{
			Src:     dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			Dst:     dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.jpg"),
			MovedBy: &users.ExampleAlice,
		}).Return(errs.ErrValidation).Once()

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			&storage.PaginateCmd{StartAfter: map[string]string{"name": ""}, Limit: PageSize}).
			Return([]dfs.INode{dfs.ExampleAliceFile2}, nil).Once()

		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg")).
			Return(&dfs.ExampleAliceFile, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, &browser.MoveTemplate{
			SrcPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/file.jpg"),
			SrcInode:      &dfs.ExampleAliceFile,
			DstPath:       dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/"),
			FolderContent: map[dfs.PathCmd]dfs.INode{*dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/bar/file.txt"): dfs.ExampleAliceFile2},
			PageSize:      PageSize,
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/browser/move", nil)
		r.URL.RawQuery = url.Values{
			"srcPath": []string{"/foo/file.jpg"},
			"dstPath": []string{"/bar/"},
			"spaceID": []string{"some-space-id"},
		}.Encode()

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})
}
