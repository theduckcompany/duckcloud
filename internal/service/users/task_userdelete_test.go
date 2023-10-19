package users

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestUserDeleteTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewUserDeleteTaskRunner(nil, nil, nil, nil, nil, nil, nil)
		assert.Equal(t, "user-delete", job.Name())
	})

	t.Run("Run success", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		webSessionsMock.On("DeleteAll", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b"), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()
		dfsMock.On("RemoveFS", mock.Anything, &folders.ExampleAlicePersonalFolder).Return(nil).Once()

		oauthConsentMock.On("DeleteAll", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()

		err := job.Run(ctx, json.RawMessage(`{"user-id": "b13c77ab-02fa-48a0-aad4-2079b6894d7b"}`))
		assert.NoError(t, err)
	})

	t.Run("Run with some invalid json", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()
		dfsMock.On("RemoveFS", mock.Anything, &folders.ExampleAlicePersonalFolder).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.NoError(t, err)
	})

	t.Run("with a websession deletion error", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all web sessions: some-error")
	})

	t.Run("with a dav session deletion error", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all dav sessions: some-error")
	})

	t.Run("with a oauth session deletion error", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all oauth sessions: some-error")
	})

	t.Run("with a rootFS deletion error", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		dfsMock.On("RemoveFS", mock.Anything, &folders.ExampleAlicePersonalFolder).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, `failed to RemoveFS: some-error`)
	})

	t.Run("with an fs deletion error", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		dfsMock.On("RemoveFS", mock.Anything, &folders.ExampleAlicePersonalFolder).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all oauth consents: some-error")
	})

	t.Run("with a user hard delete error", func(t *testing.T) {
		usersMock := NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		dfsMock := dfs.NewMockService(t)
		job := NewUserDeleteTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, dfsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		dfsMock.On("RemoveFS", mock.Anything, &folders.ExampleAlicePersonalFolder).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, ExampleDeletingAlice.ID()).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to hard delete the user: some-error")
	})
}
