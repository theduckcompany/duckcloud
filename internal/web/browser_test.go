package web

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

func Test_Browser_Page(t *testing.T) {
	t.Run("redirectDefaultBrowser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/browser/e97b60f7-add2-43e1-a9bd-e2dac9ce69ec", res.Header.Get("Location"))
	})

	t.Run("getBrowserHome with an unauthenticated user", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("getBrowserContent success with dir", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder}, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("Folder").Return(&folders.ExampleAlicePersonalFolder)

		// Then look for the path inside this folder
		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&dfs.ExampleAliceRoot, nil).Once()

		folderFSMock.On("ListDir", mock.Anything, "foo/bar", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		folderID := string(folders.ExampleAlicePersonalFolder.ID())
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
			"host":     "example.com",
			"fullPath": "foo/bar",
			"folder":   &folders.ExampleAlicePersonalFolder,
			"breadcrumb": []breadCrumbElement{
				{Name: folders.ExampleAlicePersonalFolder.Name(), Href: "/browser/" + folderID, Current: false},
				{Name: "foo", Href: "/browser/" + folderID + "/foo", Current: false},
				{Name: "bar", Href: "/browser/" + folderID + "/foo/bar", Current: true},
			},
			"folders": []folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder},
			"inodes":  []dfs.INode{dfs.ExampleAliceFile},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/folder-id/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getBrowserContent success with file", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("Folder").Return(&folders.ExampleAlicePersonalFolder)

		// Then look for the path inside this folder
		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&dfs.ExampleAliceFile, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, *dfs.ExampleAliceFile.FileID()).Return(&files.ExampleFile1, nil).Once()

		afs := afero.NewMemMapFs()
		file, err := afero.TempFile(afs, t.TempDir(), "")
		require.NoError(t, err)

		folderFSMock.On("Download", mock.Anything, "foo/bar").Return(file, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/folder-id/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getBrowserContent with an unauthenticated user", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/folder-id/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("getBrowserContent with an invalid folder id", func(t *testing.T) {
		// The url is not correctly formed. The path is missing so we
		// redirect the user to the browser home page.
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID(""), fmt.Errorf("invalid id")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/folder-id/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/browser", res.Header.Get("Location"))
	})

	t.Run("getBrowserContent with a folder not found", func(t *testing.T) {
		// The url is not correctly formed. The path is missing so we
		// redirect the user to the browser home page.
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(nil, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/folder-id/foo", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/browser", res.Header.Get("Location"))
	})

	t.Run("getBrowserContent with an invalid file path", func(t *testing.T) {
		// The url is not correctly formed. The path is missing so we
		// redirect the user to the browser home page.
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("Folder").Return(&folders.ExampleAlicePersonalFolder)

		// Then look for the path inside this folder
		folderFSMock.On("Get", mock.Anything, "invalid").Return(nil, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/folder-id/invalid", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		folderID := string(folders.ExampleAlicePersonalFolder.ID())
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/browser/"+folderID, res.Header.Get("Location"))
	})

	t.Run("upload file success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		content := "Hello, World!"

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()

		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("CreateDir", mock.Anything, "foo/bar").Return(&dfs.ExampleAliceDir, nil).Once()
		folderFSMock.On("Upload", mock.Anything, "foo/bar/hello.txt", mock.Anything).
			Run(func(args mock.Arguments) {
				uploaded, err := io.ReadAll(args[2].(io.Reader))
				require.NoError(t, err)
				require.Equal(t, []byte(content), uploaded)
			}).
			Return(nil).Once()

		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		form.WriteField("name", "hello.txt")
		form.WriteField("rootPath", "/foo/bar") // This correspond to the DuckFS path where the upload append
		form.WriteField("folderID", "folder-id")
		writer, err := form.CreateFormFile("file", "hello.txt")
		require.NoError(t, err)
		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
		require.NoError(t, form.Close())

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/browser/upload", buf)
		r.Header.Set("Content-Type", form.FormDataContentType())

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("upload folder success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		content := "Hello, World!"

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()

		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("CreateDir", mock.Anything, "foo/bar/baz").Return(&dfs.ExampleAliceDir, nil).Once()

		folderFSMock.On("Upload", mock.Anything, "foo/bar/baz/hello.txt", mock.Anything).
			Run(func(args mock.Arguments) {
				uploaded, err := io.ReadAll(args[2].(io.Reader))
				require.NoError(t, err)
				require.Equal(t, []byte(content), uploaded)
			}).
			Return(nil).Once()

		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		form.WriteField("name", "hello.txt")
		form.WriteField("rootPath", "/foo/bar")
		form.WriteField("folderID", "folder-id")
		form.WriteField("relativePath", "/baz/hello.txt")
		writer, err := form.CreateFormFile("file", "hello.txt")
		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
		require.NoError(t, form.Close())
		writer.Write([]byte(content))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/browser/upload", buf)
		r.Header.Set("Content-Type", form.FormDataContentType())

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("Folder").Return(&folders.ExampleAlicePersonalFolder)

		folderFSMock.On("Remove", mock.Anything, "foo/bar").Return(nil).Once()

		// Then look for the path inside this folder
		folderFSMock.On("Get", mock.Anything, "foo").Return(&dfs.ExampleAliceRoot, nil).Once()

		folderFSMock.On("ListDir", mock.Anything, "foo", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder}, nil).Once()

		folderID := string(folders.ExampleAlicePersonalFolder.ID())
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
			"host":     "example.com",
			"fullPath": "foo",
			"folder":   &folders.ExampleAlicePersonalFolder,
			"breadcrumb": []breadCrumbElement{
				{Name: folders.ExampleAlicePersonalFolder.Name(), Href: "/browser/" + folderID, Current: false},
				{Name: "foo", Href: "/browser/" + folderID + "/foo", Current: true},
			},
			"folders": []folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder},
			"inodes":  []dfs.INode{dfs.ExampleAliceFile},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/browser/folder-id/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getCreateDirModel success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/create-dir.tmpl", map[string]any{
			"directory": "/foo/bar",
			"folderID":  "some-folder-id",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		queries.Add("dir", "/foo/bar")
		queries.Add("folder", "some-folder-id")

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
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrSessionNotFound).Once()

		webSessionsMock.On("Logout", mock.Anything, mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		queries.Add("dir", "/foo/bar")
		queries.Add("folder", "some-folder-id")

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
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errors.New("failed to get the dir path from the url query")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		// queries.Add("dir", "/foo/bar")
		queries.Add("folder", "some-folder-id")

		r.URL.RawQuery = queries.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getCreateDirModel with no dir query", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errors.New("failed to get the folder id from the url query")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/create-dir", nil)
		queries := url.Values{}
		queries.Add("dir", "/foo/bar")
		// queries.Add("folder", "some-folder-id")

		r.URL.RawQuery = queries.Encode()
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-folder-id").Return(uuid.UUID("some-folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("some-folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("Get", mock.Anything, "/foo/New Dir").Return(nil, errs.ErrNotFound).Once()
		folderFSMock.On("CreateDir", mock.Anything, "/foo/New Dir").Return(&dfs.ExampleAliceDir, nil).Once()

		// Render
		folderFSMock.On("Get", mock.Anything, "/foo").Return(&dfs.ExampleAliceRoot, nil).Once()
		folderFSMock.On("Folder").Return(&folders.ExampleAlicePersonalFolder).Once()
		folderFSMock.On("ListDir", mock.Anything, "/foo", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		}).Return([]dfs.INode{dfs.ExampleAliceFile, dfs.ExampleAliceFile2}, nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder}, nil).Once()
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/content.tmpl", map[string]any{
			"breadcrumb": []breadCrumbElement{
				{
					Name:    "Alice's Folder",
					Href:    "/browser/e97b60f7-add2-43e1-a9bd-e2dac9ce69ec",
					Current: false,
				},
				{
					Name:    "",
					Href:    "/browser/e97b60f7-add2-43e1-a9bd-e2dac9ce69ec",
					Current: false,
				},
				{
					Name:    "foo",
					Href:    "/browser/e97b60f7-add2-43e1-a9bd-e2dac9ce69ec/foo",
					Current: true,
				},
			},
			"folder":   &folders.ExampleAlicePersonalFolder,
			"folders":  []folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder},
			"fullPath": "/foo",
			"host":     "example.com",
			"inodes":   []dfs.INode{dfs.ExampleAliceFile, dfs.ExampleAliceFile2},
		})

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("folderID", "some-folder-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with an authentication error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("folderID", "some-folder-id")
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

	t.Run("handleCreateDirReq with an invalid folderID", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-folder-id").Return(uuid.UUID(""), fmt.Errorf("some-error")).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, errors.New("invalid folder id param")).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		// form.Add("name", "New Dir")
		form.Add("folderID", "some-folder-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with an empty name", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-folder-id").Return(uuid.UUID("some-folder-id"), nil).Once()
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]any{
			"directory": "/foo",
			"folderID":  uuid.UUID("some-folder-id"),
			"error":     "Must not be empty",
		}).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		// form.Add("name", "New Dir")
		form.Add("folderID", "some-folder-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with a taken name", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-folder-id").Return(uuid.UUID("some-folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("some-folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("Get", mock.Anything, "/foo/New Dir").Return(&dfs.ExampleAliceDir, nil).Once()

		// Render
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]any{
			"directory": "/foo",
			"folderID":  uuid.UUID("some-folder-id"),
			"error":     "Already exists",
		}).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("folderID", "some-folder-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("handleCreateDirReq with a CreateDir error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-folder-id").Return(uuid.UUID("some-folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("some-folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		err := fmt.Errorf("some-error")
		folderFSMock.On("Get", mock.Anything, "/foo/New Dir").Return(nil, errs.ErrNotFound).Once()
		folderFSMock.On("CreateDir", mock.Anything, "/foo/New Dir").Return(nil, err).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to create the directory: %w", err)).Once()

		w := httptest.NewRecorder()
		form := url.Values{}
		form.Add("dirPath", "/foo")
		form.Add("name", "New Dir")
		form.Add("folderID", "some-folder-id")
		r := httptest.NewRequest(http.MethodPost, "/browser/create-dir", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})
}
