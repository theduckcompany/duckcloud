package usercreate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestUserCreateJob(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, inodesMock, foldersMock, tools)

		usersMock.On("GetAllWithStatus", mock.Anything, "initializing", &storage.PaginateCmd{Limit: batchSize}).
			Return([]users.User{users.ExampleInitializingAlice}, nil).Once()
		foldersMock.On("CreatePersonalFolder", mock.Anything, &folders.CreatePersonalFolderCmd{
			Name:  "My files",
			Owner: users.ExampleAlice.ID(),
		}).Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		usersMock.On("SetDefaultFolder", mock.Anything, users.ExampleInitializingAlice, &folders.ExampleAlicePersonalFolder).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("MarkInitAsFinished", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		err := job.Run(ctx)
		assert.NoError(t, err)
	})
}
