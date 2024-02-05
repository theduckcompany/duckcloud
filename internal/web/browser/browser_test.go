package browser

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
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

func Test_Browser_Page(t *testing.T) {
	t.Run("redirectDefaultBrowser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		spacesMock.On("GetAllUserSpaces", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]spaces.Space{spaces.ExampleAlicePersonalSpace}, nil)

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
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

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
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		spacesMock.On("GetAllUserSpaces", mock.Anything, users.ExampleAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleAliceBobSharedSpace}, nil).Once()

		// Get the space from the url
		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		// Then look for the path inside this space
		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar")).Return(&dfs.ExampleAliceRoot, nil).Once()

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"), &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &browser.ContentTemplate{
			Folder:        dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"),
			Inodes:        []dfs.INode{dfs.ExampleAliceFile},
			CurrentSpace:  &spaces.ExampleAlicePersonalSpace,
			AllSpaces:     []spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleAliceBobSharedSpace},
			ContentTarget: "#content",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, path.Join("/browser", string(spaces.ExampleAlicePersonalSpace.ID()), "/foo/bar"), w.Header().Get("HX-Push-Url"))
	})

	t.Run("getBrowserContent success with dir and a last attribute", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space from the url
		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("ListDir", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"), &storage.PaginateCmd{
			StartAfter: map[string]string{"name": "some-filename"},
			Limit:      PageSize,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &browser.RowsTemplate{
			Folder:        dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"),
			Inodes:        []dfs.INode{dfs.ExampleAliceFile},
			ContentTarget: "#content",
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo/bar?last=some-filename", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Empty(t, w.Header().Get("HX-Push-Url"))
	})

	t.Run("getBrowserContent success with file", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space from the url
		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		// Then look for the path inside this space
		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar")).Return(&dfs.ExampleAliceFile, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, *dfs.ExampleAliceFile.FileID()).Return(&files.ExampleFile1, nil).Once()

		afs := afero.NewMemMapFs()
		file, err := afero.TempFile(afs, t.TempDir(), "")
		require.NoError(t, err)

		fsMock.On("Download", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar")).Return(file, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo/bar", nil)
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
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("getBrowserContent with an invalid space id", func(t *testing.T) {
		// The url is not correctly formed. The path is missing so we
		// redirect the user to the browser home page.
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID(""), fmt.Errorf("invalid id")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/browser", res.Header.Get("Location"))
	})

	t.Run("getBrowserContent with a space not found", func(t *testing.T) {
		// The url is not correctly formed. The path is missing so we
		// redirect the user to the browser home page.
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space from the url
		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(nil, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo", nil)
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
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space from the url
		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		// Then look for the path inside this space
		fsMock.On("Get", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/invalid")).Return(nil, errs.ErrNotFound).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/invalid", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		spaceID := string(spaces.ExampleAlicePersonalSpace.ID())
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/browser/"+spaceID, res.Header.Get("Location"))
	})

	t.Run("upload file success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		content := "Hello, World!"

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()

		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("CreateDir", mock.Anything, &dfs.CreateDirCmd{
			Path:      dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "foo/bar"),
			CreatedBy: &users.ExampleAlice,
		}).Return(&dfs.ExampleAliceDir, nil).Once()
		fsMock.On("Upload", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				cmd, ok := args[1].(*dfs.UploadCmd)
				require.True(t, ok)

				require.Equal(t, "foo/bar/hello.txt", cmd.FilePath)
				require.Equal(t, &users.ExampleAlice, cmd.UploadedBy)

				uploaded, err := io.ReadAll(cmd.Content)
				require.NoError(t, err)
				require.Equal(t, []byte(content), uploaded)
			}).
			Return(nil).Once()

		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		form.WriteField("name", "hello.txt")
		form.WriteField("rootPath", "/foo/bar") // This correspond to the DuckFS path where the upload append
		form.WriteField("spaceID", "d09f29f9-5131-4aa4-b69c-7717124b213e")
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

	t.Run("upload space success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		content := "Hello, World!"

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()

		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("CreateDir", mock.Anything, &dfs.CreateDirCmd{
			Path:      dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "foo/bar/baz"),
			CreatedBy: &users.ExampleAlice,
		}).Return(&dfs.ExampleAliceDir, nil).Once()

		fsMock.On("Upload", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				cmd, ok := args[1].(*dfs.UploadCmd)
				require.True(t, ok)

				require.Equal(t, "foo/bar/baz/hello.txt", cmd.FilePath)
				require.Equal(t, &users.ExampleAlice, cmd.UploadedBy)

				uploaded, err := io.ReadAll(cmd.Content)
				require.NoError(t, err)
				require.Equal(t, []byte(content), uploaded)
			}).
			Return(nil).Once()

		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		form.WriteField("name", "hello.txt")
		form.WriteField("rootPath", "/foo/bar")
		form.WriteField("spaceID", "d09f29f9-5131-4aa4-b69c-7717124b213e")
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
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		filesMock := files.NewMockService(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		fsMock := dfs.NewMockService(t)
		handler := NewHandler(tools, htmlMock, spacesMock, filesMock, auth, fsMock)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Get the space from the url
		tools.UUIDMock.On("Parse", "d09f29f9-5131-4aa4-b69c-7717124b213e").Return(uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("d09f29f9-5131-4aa4-b69c-7717124b213e")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		fsMock.On("Remove", mock.Anything, dfs.NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar")).Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/browser/d09f29f9-5131-4aa4-b69c-7717124b213e/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Empty(t, body)
	})
}
