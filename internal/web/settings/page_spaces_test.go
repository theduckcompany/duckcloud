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
	t.Parallel()

	t.Run("getContent success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		user2 := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).CreatedBy(user).Build()
		space2 := spaces.NewFakeSpace(t).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).Return([]users.User{*user, *user2}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, user, (*sqlstorage.PaginateCmd)(nil)).Return([]spaces.Space{*space, *space2}, nil).Once()
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &spacestmpl.ContentTemplate{
			IsAdmin: true,
			Spaces:  []spaces.Space{*space, *space2},
			Users:   map[uuid.UUID]users.User{user.ID(): *user, user2.ID(): *user2},
		})

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getContent with an authentication error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		schedulerMock := scheduler.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(nil, websessions.ErrMissingSessionToken).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))
	})

	t.Run("getContent with a non admin user", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).Build() // NOTE: is not an admin
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		assert.Equal(t, "/settings", res.Header.Get("Location"))
	})

	t.Run("getContent with a users.GetAll error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).Return(nil, errs.ErrInternal).Once()

		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to GetAllUsers: %w", errs.ErrInternal))

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getContent with a spaces.GetAllSpaces error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).Return([]users.User{*user}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, user, (*sqlstorage.PaginateCmd)(nil)).
			Return(nil, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to GetAllSpaces: %w", errs.ErrInternal))

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteSpace success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		someSpaceID := "some-space-id"

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", someSpaceID).Return(uuid.UUID(someSpaceID), nil).Once()
		spacesMock.On("Delete", mock.Anything, user, uuid.UUID(someSpaceID)).Return(nil).Once()

		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).
			Return([]users.User{*user}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, user, (*sqlstorage.PaginateCmd)(nil)).
			Return(nil, errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to GetAllSpaces: %w", errs.ErrInternal))

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/"+someSpaceID+"/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("deleteSpace with an invalid uuid inside the url", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", "some-invalid-id").Return(uuid.UUID(""), errs.ErrValidation).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/some-invalid-id/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("deleteSpace with a delete error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()
		someSpaceID := "some-space-id"

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", someSpaceID).Return(uuid.UUID(someSpaceID), nil).Once()
		spacesMock.On("Delete", mock.Anything, user, uuid.UUID(someSpaceID)).Return(errs.ErrInternal).Once()
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to Delete the space: %w", errs.ErrInternal))

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/"+someSpaceID+"/delete", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getCreateSpaceModal success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		user2 := users.NewFakeUser(t).Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).
			Return([]users.User{*user, *user2}, nil)
		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &spacestmpl.CreateSpaceModal{
			IsAdmin: user.IsAdmin(),
			Selection: spacestmpl.UserSelectionTemplate{
				UnselectedUsers: []users.User{*user, *user2},
				SelectedUsers:   []users.User{*user},
			},
		})

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces/new", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Asserts
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("getCreateSpaceModal with an non admin user", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).Build() // NOTE: Not an admin
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces/new", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Assert
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("getCreateSpaceModal with a users.GetAll error", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).Return(nil, errs.ErrBadRequest)
		htmlMock.On("WriteHTMLErrorPage", mock.Anything, mock.Anything, fmt.Errorf("failed to Get all the users: %w", errs.ErrBadRequest))

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/spaces/new", nil)
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Assert
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("createSpace success", func(t *testing.T) {
		t.Parallel()

		tools := tools.NewMock(t)
		webSessionsMock := websessions.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		usersMock := users.NewMockService(t)
		htmlMock := html.NewMockWriter(t)
		auth := auth.NewAuthenticator(webSessionsMock, usersMock, htmlMock)
		schedulerMock := scheduler.NewMockService(t)
		handler := NewSpacesPage(htmlMock, spacesMock, usersMock, auth, schedulerMock, tools)

		// Data
		user := users.NewFakeUser(t).WithAdminRole().Build()
		user2 := users.NewFakeUser(t).Build()
		space := spaces.NewFakeSpace(t).
			WithName("some-space-name").
			WithOwners(*user, *user2).
			Build()
		webSession := websessions.NewFakeSession(t).CreatedBy(user).Build()

		// Mocks
		webSessionsMock.On("GetFromReq", mock.Anything, mock.Anything).Return(webSession, nil).Once()
		usersMock.On("GetByID", mock.Anything, user.ID()).Return(user, nil).Once()
		tools.UUIDMock.On("Parse", string(user.ID())).Return(user.ID(), nil).Once()
		tools.UUIDMock.On("Parse", string(user2.ID())).Return(user2.ID(), nil).Once()
		schedulerMock.On("RegisterSpaceCreateTask", mock.Anything, &scheduler.SpaceCreateArgs{
			UserID: user.ID(),
			Name:   "some-space-name",
			Owners: []uuid.UUID{user.ID(), user2.ID()},
		}).Return(nil).Once()

		// Render the page
		usersMock.On("GetAll", mock.Anything, (*sqlstorage.PaginateCmd)(nil)).Return([]users.User{*user, *user2}, nil).Once()
		spacesMock.On("GetAllSpaces", mock.Anything, user, (*sqlstorage.PaginateCmd)(nil)).Return([]spaces.Space{*space}, nil).Once()

		htmlMock.On("WriteHTMLTemplate", mock.Anything, mock.Anything, http.StatusOK, &spacestmpl.ContentTemplate{
			IsAdmin: true,
			Spaces:  []spaces.Space{*space},
			Users: map[uuid.UUID]users.User{
				user.ID():  *user,
				user2.ID(): *user2,
			},
		})

		// Run
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/settings/spaces/create", strings.NewReader(url.Values{
			"selectedUsers": []string{string(user.ID()), string(user2.ID())},
			"name":          []string{"some-space-name"},
		}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		srv := chi.NewRouter()
		handler.Register(srv, nil)
		srv.ServeHTTP(w, r)

		// Assert
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
