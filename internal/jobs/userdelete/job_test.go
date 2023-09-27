package userdelete

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestUserDeleteJob(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)

		folderFSMock.On("RemoveAll", mock.Anything, "/").Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()

		oauthConsentMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()

		err := job.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("with a GetAllWithStatus error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return(nil, fmt.Errorf("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to GetAllWithStatus: some-error")
	})

	t.Run("with a websession deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to delete user \"86bffce3-3f53-4631-baf8-8530773884f3\": failed to delete all web sessions: some-error")
	})

	t.Run("with a dav session deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to delete user \"86bffce3-3f53-4631-baf8-8530773884f3\": failed to delete all dav sessions: some-error")
	})

	t.Run("with a oauth session deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to delete user \"86bffce3-3f53-4631-baf8-8530773884f3\": failed to delete all oauth sessions: some-error")
	})

	t.Run("with a fs deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("RemoveAll", mock.Anything, "/").Return(errors.New("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to delete user \"86bffce3-3f53-4631-baf8-8530773884f3\": failed to delete the user root fs: some-error")
	})

	t.Run("with an fs deletion error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("RemoveAll", mock.Anything, "/").Return(nil).Once()

		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to delete user \"86bffce3-3f53-4631-baf8-8530773884f3\": failed to delete all oauth consents: some-error")
	})

	t.Run("with a user hard delete error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		webSessionsMock := websessions.NewMockService(t)
		davSessionsMock := davsessions.NewMockService(t)
		oauthSessionsMock := oauthsessions.NewMockService(t)
		oauthConsentMock := oauthconsents.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		fsMock := fs.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, webSessionsMock, davSessionsMock, oauthSessionsMock, oauthConsentMock, foldersMock, fsMock, tools)

		// Fetch all the users to delete
		usersMock.On("GetAllWithStatus", mock.Anything, "deleting", &storage.PaginateCmd{Limit: gcBatchSize}).Return([]users.User{users.ExampleDeletingAlice}, nil).Once()

		// For each users remove all the data
		webSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		davSessionsMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		oauthSessionsMock.On("DeleteAllForUser", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		foldersMock.On("GetAllUserFolders", mock.Anything, users.ExampleDeletingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]folders.Folder{folders.ExampleAlicePersonalFolder}, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock)
		folderFSMock.On("RemoveAll", mock.Anything, "/").Return(nil).Once()
		foldersMock.On("Delete", mock.Anything, folders.ExampleAlicePersonalFolder.ID()).Return(nil).Once()
		oauthConsentMock.On("DeleteAll", mock.Anything, users.ExampleDeletingAlice.ID()).Return(nil).Once()
		usersMock.On("HardDelete", mock.Anything, users.ExampleDeletingAlice.ID()).Return(fmt.Errorf("some-error")).Once()

		err := job.Run(ctx)
		assert.EqualError(t, err, "failed to delete user \"86bffce3-3f53-4631-baf8-8530773884f3\": failed to hard delete the user: some-error")
	})
}
