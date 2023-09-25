package web

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder}, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		// Then look for the path inside this folder
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo/bar",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		inodesMock.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo/bar",
		}, (*storage.PaginateCmd)(nil)).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		folderID := string(folders.ExampleAlicePersonalFolder.ID())
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
			"fullPath": "foo/bar",
			"folder":   &folders.ExampleAlicePersonalFolder,
			"breadcrumb": []breadCrumbElement{
				{Name: folders.ExampleAlicePersonalFolder.Name(), Href: "/browser/" + folderID, Current: false},
				{Name: "foo", Href: "/browser/" + folderID + "/foo", Current: false},
				{Name: "bar", Href: "/browser/" + folderID + "/foo/bar", Current: true},
			},
			"folders": []folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder},
			"inodes":  []inodes.INode{inodes.ExampleAliceFile},
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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		// Then look for the path inside this folder
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo/bar",
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		afs := afero.NewMemMapFs()
		file, err := afero.TempFile(afs, t.TempDir(), "")
		require.NoError(t, err)

		filesMock.On("Open", mock.Anything, inodes.ExampleAliceFile.ID()).Return(file, nil).Once()

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		// Then look for the path inside this folder
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "invalid",
		}).Return(nil, nil).Once()

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()

		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		fileOrDir := fs.NewMockFileOrDirectory(t)
		folderFSMock.On("OpenFile", mock.Anything, "foo/bar/hello.txt", os.O_CREATE|os.O_EXCL|os.O_WRONLY).Return(fileOrDir, nil).Once()

		fileOrDir.On("Write", []byte("Hello, World!")).Return(13, nil)
		fileOrDir.On("Close").Return(nil)

		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		form.WriteField("name", "hello.txt")
		form.WriteField("rootPath", "/foo/bar") // This correspond to the DuckFS path where the upload append
		form.WriteField("folderID", "folder-id")
		writer, err := form.CreateFormFile("file", "hello.txt")
		require.NoError(t, err)
		writer.Write([]byte("Hello, World!"))

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()

		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("CreateDir", mock.Anything, "/foo/bar/baz").Return(&inodes.ExampleAliceRoot, nil).Once()

		fileOrDir := fs.NewMockFileOrDirectory(t)
		folderFSMock.On("OpenFile", mock.Anything, "foo/bar/baz/hello.txt", os.O_CREATE|os.O_EXCL|os.O_WRONLY).Return(fileOrDir, nil).Once()

		fileOrDir.On("Write", []byte("Hello, World!")).Return(13, nil)
		fileOrDir.On("Close").Return(nil)

		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		form.WriteField("name", "hello.txt")
		form.WriteField("rootPath", "/foo/bar")
		form.WriteField("folderID", "folder-id")
		form.WriteField("relativePath", "/baz/hello.txt")
		writer, err := form.CreateFormFile("file", "hello.txt")
		require.NoError(t, err)
		writer.Write([]byte("Hello, World!"))

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
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := fs.NewMockService(t)
		handler := newBrowserHandler(tools, htmlMock, foldersMock, inodesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the folder from the url
		tools.UUIDMock.On("Parse", "folder-id").Return(uuid.UUID("folder-id"), nil).Once()
		foldersMock.On("GetUserFolder", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("folder-id")).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("RemoveAll", mock.Anything, "foo/bar").Return(nil).Once()

		// Then look for the path inside this folder
		inodesMock.On("Get", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo",
		}).Return(&inodes.ExampleAliceRoot, nil).Once()

		inodesMock.On("Readdir", mock.Anything, &inodes.PathCmd{
			Root:     folders.ExampleAlicePersonalFolder.RootFS(),
			FullName: "foo",
		}, (*storage.PaginateCmd)(nil)).Return([]inodes.INode{inodes.ExampleAliceFile}, nil).Once()

		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder}, nil).Once()

		folderID := string(folders.ExampleAlicePersonalFolder.ID())
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
			"fullPath": "foo",
			"folder":   &folders.ExampleAlicePersonalFolder,
			"breadcrumb": []breadCrumbElement{
				{Name: folders.ExampleAlicePersonalFolder.Name(), Href: "/browser/" + folderID, Current: false},
				{Name: "foo", Href: "/browser/" + folderID + "/foo", Current: true},
			},
			"folders": []folders.Folder{folders.ExampleAlicePersonalFolder, folders.ExampleAliceBobSharedFolder},
			"inodes":  []inodes.INode{inodes.ExampleAliceFile},
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
}
