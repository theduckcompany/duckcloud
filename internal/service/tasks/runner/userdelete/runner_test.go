package userdelete

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestUserDeleteTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewTaskRunner(nil, nil, nil, nil, nil, nil, nil)
		assert.Equal(t, model.UserDelete, job.Name())
	})

	t.Run("Run success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		webSessionsMock.On("DeleteAll", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b"), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("Remove", mock.Anything, "/").Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		oauthConsentMock.On("DeleteAll", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, uuid.UUID("b13c77ab-02fa-48a0-aad4-2079b6894d7b")).Return(nil).Once()

		err := job.Run(ctx, json.RawMessage(`{"user-id": "b13c77ab-02fa-48a0-aad4-2079b6894d7b"}`))
		assert.NoError(t, err)
	})

	t.Run("Run with some invalid json", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("Remove", mock.Anything, "/").Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		oauthConsentMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.NoError(t, err)
	})

	t.Run("with a websession deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all web sessions: some-error")
	})

	t.Run("with a dav session deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all dav sessions: some-error")
	})

	t.Run("with a oauth session deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all oauth sessions: some-error")
	})

	t.Run("with a fs deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("Remove", mock.Anything, "/").Return(errors.New("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete the user root fs: some-error")
	})

	t.Run("with an fs deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("Remove", mock.Anything, "/").Return(nil).Once()

		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to delete all oauth consents: some-error")
	})

	t.Run("with a user hard delete error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := dfs.NewMockService(t)

		job := NewTaskRunner(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock)

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := dfs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("Remove", mock.Anything, "/").Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.RunArgs(ctx, &scheduler.UserDeleteArgs{UserID: users.ExampleDeletingAlice.ID()})
		assert.EqualError(t, err, "failed to hard delete the user: some-error")
	})
}
