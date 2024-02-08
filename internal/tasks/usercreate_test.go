package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestUserCreateTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewUserCreateTaskRunner(nil, nil, nil)
		assert.Equal(t, "user-create", job.Name())
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).Return(&users.ExampleInitializingBob, nil).Once()
		usersMock.On("MarkInitAsFinished", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).Return(&users.ExampleBob, nil).Once()

		err := job.Run(ctx, json.RawMessage(`{"user-id": "059d78af-e675-498e-8b77-d4b2b4b9d4e7"}`))
		require.NoError(t, err)
	})

	t.Run("RunArgs with some invalid json args", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		require.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("with an already active user", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewUserCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, users.ExampleBob.ID()).Return(&users.ExampleBob, nil).Once()

		// Do nothing

		err := job.RunArgs(ctx, &scheduler.UserCreateArgs{UserID: users.ExampleBob.ID()})
		require.NoError(t, err)
	})
}
