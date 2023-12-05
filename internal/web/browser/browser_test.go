package browser

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		spaceFSMock.On("Space").Return(&spaces.ExampleAlicePersonalSpace)

		// Then look for the path inside this space
		spaceFSMock.On("Get", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"}).Return(&dfs.ExampleAliceRoot, nil).Once()

		spaceFSMock.On("ListDir", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"}, &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		spaceID := string(spaces.ExampleAlicePersonalSpace.ID())
		htmlMock.On("WriteHTML", mock.Anything, mock.Anything, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
			"host":     "example.com",
			"fullPath": "/foo/bar",
			"space":    &spaces.ExampleAlicePersonalSpace,
			"breadcrumb": []breadCrumbElement{
				{Name: spaces.ExampleAlicePersonalSpace.Name(), Href: "/browser/" + spaceID, Current: false},
				{Name: "foo", Href: "/browser/" + spaceID + "/foo", Current: false},
				{Name: "bar", Href: "/browser/" + spaceID + "/foo/bar", Current: true},
			},
			"spaces": []spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleAliceBobSharedSpace},
			"inodes": []dfs.INode{dfs.ExampleAliceFile},
		}).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/space-id/foo/bar", nil)
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
		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)
		spaceFSMock.On("Space").Return(&spaces.ExampleAlicePersonalSpace)

		// Then look for the path inside this space
		spaceFSMock.On("Get", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"}).Return(&dfs.ExampleAliceFile, nil).Once()

		filesMock.On("GetMetadata", mock.Anything, *dfs.ExampleAliceFile.FileID()).Return(&files.ExampleFile1, nil).Once()

		afs := afero.NewMemMapFs()
		file, err := afero.TempFile(afs, t.TempDir(), "")
		require.NoError(t, err)

		spaceFSMock.On("Download", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"}).Return(file, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/space-id/foo/bar", nil)
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
		r := httptest.NewRequest(http.MethodGet, "/browser/space-id/foo/bar", nil)
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

		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID(""), fmt.Errorf("invalid id")).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/space-id/foo/bar", nil)
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
		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(nil, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/space-id/foo", nil)
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
		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)
		spaceFSMock.On("Space").Return(&spaces.ExampleAlicePersonalSpace)

		// Then look for the path inside this space
		spaceFSMock.On("Get", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/invalid"}).Return(nil, errs.ErrNotFound).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/browser/space-id/invalid", nil)
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

		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()

		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		spaceFSMock.On("CreateDir", mock.Anything, &dfs.CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "foo/bar",
			CreatedBy: &users.ExampleAlice,
		}).Return(&dfs.ExampleAliceDir, nil).Once()
		spaceFSMock.On("Upload", mock.Anything, mock.Anything).
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
		form.WriteField("spaceID", "space-id")
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

		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()

		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		spaceFSMock.On("CreateDir", mock.Anything, &dfs.CreateDirCmd{
			Space:     &spaces.ExampleAlicePersonalSpace,
			FilePath:  "foo/bar/baz",
			CreatedBy: &users.ExampleAlice,
		}).Return(&dfs.ExampleAliceDir, nil).Once()

		spaceFSMock.On("Upload", mock.Anything, mock.Anything).
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
		form.WriteField("spaceID", "space-id")
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
		tools.UUIDMock.On("Parse", "space-id").Return(uuid.UUID("space-id"), nil).Once()
		spacesMock.On("GetUserSpace", mock.Anything, users.ExampleAlice.ID(), uuid.UUID("space-id")).
			Return(&spaces.ExampleAlicePersonalSpace, nil).Once()

		spaceFSMock := dfs.NewMockFS(t)
		fsMock.On("GetSpaceFS", &spaces.ExampleAlicePersonalSpace).Return(spaceFSMock)

		spaceFSMock.On("Remove", mock.Anything, &dfs.PathCmd{Space: &spaces.ExampleAlicePersonalSpace, Path: "/foo/bar"}).Return(nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/browser/space-id/foo/bar", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Equal(t, "refreshFolder", res.Header.Get("HX-Trigger"))
	})
}
