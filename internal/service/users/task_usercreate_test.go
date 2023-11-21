package users

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestUserCreateTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewUserCreateTaskRunner(nil, nil, nil)
		assert.Equal(t, "user-create", job.Name())
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).Return(&ExampleInitializingAlice, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, ExampleInitializingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]spaces.Space{}, nil).Once()
		fsMock.On("CreateFS", mock.Anything, []uuid.UUID{ExampleInitializingAlice.ID()}).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		usersMock.On("SetDefaultSpace", mock.Anything, ExampleInitializingAlice, &spaces.ExampleAlicePersonalSpace).Return(&ExampleAlice, nil).Once()

		usersMock.On("MarkInitAsFinished", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).Return(&ExampleAlice, nil).Once()

		err := job.Run(ctx, json.RawMessage(`{"user-id": "059d78af-e675-498e-8b77-d4b2b4b9d4e7"}`))
		assert.NoError(t, err)
	})

	t.Run("RunArgs with some invalid json args", func(t *testing.T) {
		usersMock := NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("RunArgs with a space already set", func(t *testing.T) {
		usersMock := NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleInitializingAlice, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, ExampleInitializingAlice.ID(), (*storage.PaginateCmd)(nil)).Return([]spaces.Space{spaces.ExampleAlicePersonalSpace}, nil).Once()
		usersMock.On("SetDefaultSpace", mock.Anything, ExampleInitializingAlice, &spaces.ExampleAlicePersonalSpace).Return(&ExampleAlice, nil).Once()

		usersMock.On("MarkInitAsFinished", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		err := job.RunArgs(ctx, &scheduler.UserCreateArgs{UserID: ExampleAlice.ID()})
		assert.NoError(t, err)
	})

	t.Run("RunArgs with a user already owning several spaces", func(t *testing.T) {
		usersMock := NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleInitializingAlice, nil).Once()
		spacesMock.On("GetAllUserSpaces", mock.Anything, ExampleInitializingAlice.ID(), (*storage.PaginateCmd)(nil)).
			Return([]spaces.Space{spaces.ExampleAlicePersonalSpace, spaces.ExampleAliceBobSharedSpace}, nil).Once()

		err := job.RunArgs(ctx, &scheduler.UserCreateArgs{UserID: ExampleAlice.ID()})
		assert.ErrorContains(t, err, "the new user already have several space")
	})

	t.Run("with an already active user", func(t *testing.T) {
		usersMock := NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, ExampleAlice.ID()).Return(&ExampleAlice, nil).Once()

		// Do nothing

		err := job.RunArgs(ctx, &scheduler.UserCreateArgs{UserID: ExampleAlice.ID()})
		assert.NoError(t, err)
	})
}
