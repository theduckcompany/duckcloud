package usercreate

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestUserCreateTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewTaskRunner(nil, nil)
		assert.Equal(t, model.UserCreate, job.Name())
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		job := NewTaskRunner(usersMock, foldersMock)

		usersMock.On("GetByID", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).Return(&users.ExampleInitializingAlice, nil).Once()

		foldersMock.On("CreatePersonalFolder", mock.Anything, &folders.CreatePersonalFolderCmd{
			Name:  "My files",
			Owner: uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7"),
		}).Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		usersMock.On("SetDefaultFolder", mock.Anything, users.ExampleInitializingAlice, &folders.ExampleAlicePersonalFolder).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("MarkInitAsFinished", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).Return(&users.ExampleAlice, nil).Once()

		err := job.Run(ctx, json.RawMessage(`{"user-id": "059d78af-e675-498e-8b77-d4b2b4b9d4e7"}`))
		assert.NoError(t, err)
	})

	t.Run("RunArgs with some invalid json args", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		job := NewTaskRunner(usersMock, foldersMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		foldersMock := folders.NewMockService(t)
		job := NewTaskRunner(usersMock, foldersMock)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleInitializingAlice, nil).Once()

		foldersMock.On("CreatePersonalFolder", mock.Anything, &folders.CreatePersonalFolderCmd{
			Name:  "My files",
			Owner: users.ExampleAlice.ID(),
		}).Return(&folders.ExampleAlicePersonalFolder, nil).Once()
		usersMock.On("SetDefaultFolder", mock.Anything, users.ExampleInitializingAlice, &folders.ExampleAlicePersonalFolder).Return(&users.ExampleAlice, nil).Once()

		usersMock.On("MarkInitAsFinished", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		err := job.RunArgs(ctx, &scheduler.UserCreateArgs{UserID: users.ExampleAlice.ID()})
		assert.NoError(t, err)
	})

	t.Run("with an already active user", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		foldersMock := folders.NewMockService(t)

		job := NewTaskRunner(usersMock, foldersMock)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).Return(&users.ExampleAlice, nil).Once()

		// Do nothing

		err := job.RunArgs(ctx, &scheduler.UserCreateArgs{UserID: users.ExampleAlice.ID()})
		assert.NoError(t, err)
	})
}
