package settings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	spacestmpl "github.com/theduckcompany/duckcloud/internal/web/html/templates/settings/spaces"
)

func Test_SpacesPage(t *testing.T) {
	t.Run("getContent success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).
			Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, &users.ExampleAlice, (*storage.PaginateCmd)(nil)).
			Return([]spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleBobPersonalSpace}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &spacestmpl.ContentTemplate{
			IsAdmin: true,
			Spaces:  []spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleBobPersonalSpace},
			Users: map[uuid.UUID]users.User{
				users.ExampleAlice.ID(): users.ExampleAlice,
				users.ExampleBob.ID():   users.ExampleBob,
			},
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getContent with an authentication error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("getContent with a non admin user", func(t *testing.T) {
		// GetByID return the user Bob which is not an admin.

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		require.False(t, users.ExampleBob.IsAdmin())

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleBob, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		assert.Equal(t, "/settings", res.Header.Get("Location"))
	})

	t.Run("getContent with a users.GetAll error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).Return(nil, errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to GetAllUsers: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getContent with a spaces.GetAllSpaces error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Render content
		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).
			Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, &users.ExampleAlice, (*storage.PaginateCmd)(nil)).
			Return(nil, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to GetAllSpaces: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("deleteSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-id").Return(uuid.UUID("some-id"), nil).Once()
		spacesMock.On("Delete", mock.Anything, &users.ExampleAlice, uuid.UUID("some-id")).Return(nil).Once()

		// Render content
		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).
			Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, &users.ExampleAlice, (*storage.PaginateCmd)(nil)).
			Return(nil, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to GetAllSpaces: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/some-id/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("deleteSpace with an invalid uuid inside the url", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-id").Return(uuid.UUID(""), errs.ErrValidation).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/some-id/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("deleteSpace with a delete error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", "some-id").Return(uuid.UUID("some-id"), nil).Once()
		spacesMock.On("Delete", mock.Anything, &users.ExampleAlice, uuid.UUID("some-id")).Return(errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to Delete the space: %w", errs.ErrInternal))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/some-id/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getCreateSpaceModal success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).
			Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil)

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &spacestmpl.CreateSpaceModal{
			IsAdmin: users.ExampleAlice.IsAdmin(),
			Selection: spacestmpl.UserSelectionTemplate{
				UnselectedUsers: []users.User{users.ExampleAlice, users.ExampleBob},
				SelectedUsers:   []users.User{users.ExampleAlice},
			},
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces/new", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("getCreateSpaceModal with an non admin user", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.BobWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleBob.ID()).Return(&users.ExampleBob, nil).Once()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces/new", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("getCreateSpaceModal with a users.GetAll error", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).
			Return(nil, errs.ErrBadRequest)

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to Get all the users: %w", errs.ErrBadRequest))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces/new", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})

	t.Run("createSpace success", func(t *testing.T) {
		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Authentication
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(&websessions.AliceWebSessionExample, nil).Once()
		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		tools.UUIDMock.On("Parse", string(users.ExampleAlice.ID())).Return(users.ExampleAlice.ID(), nil).Once()
		tools.UUIDMock.On("Parse", string(users.ExampleBob.ID())).Return(users.ExampleBob.ID(), nil).Once()
		schedulerMock.On("RegisterSpaceCreateTask", mock.Anything, &scheduler.SpaceCreateArgs{
			UserID: users.ExampleAlice.ID(),
			Name:   "some-space-name",
			Owners: []uuid.UUID{users.ExampleAlice.ID(), users.ExampleBob.ID()},
		}).Return(nil).Once()

		// Render the page
		usersMock.On("GetAll", mock.Anything, (*storage.PaginateCmd)(nil)).
			Return([]users.User{users.ExampleAlice, users.ExampleBob}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, &users.ExampleAlice, (*storage.PaginateCmd)(nil)).
			Return([]spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleBobPersonalSpace}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &spacestmpl.ContentTemplate{
			IsAdmin: true,
			Spaces:  []spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleBobPersonalSpace},
			Users: map[uuid.UUID]users.User{
				users.ExampleAlice.ID(): users.ExampleAlice,
				users.ExampleBob.ID():   users.ExampleBob,
			},
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/create", strings.NewReader(url.Values{
			"selectedUsers": []string{string(users.ExampleAlice.ID()), string(users.ExampleBob.ID())},
			"name":          []string{"some-space-name"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)
	})
}
