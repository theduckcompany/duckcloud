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
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSpaceCreateTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewSpaceCreateTaskRunner(nil, nil, nil)
		assert.Equal(t, "space-create", job.Name())
	})

	t.Run("Run with an invalid json", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewSpaceCreateTaskRunner(usersMock, spacesMock, fsMock)

		err := job.Run(ctx, json.RawMessage(`{some invalid json}`))
		require.EqualError(t, err, "failed to unmarshal the args: invalid character 's' looking for beginning of object key string")
	})

	t.Run("RunArgs success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewSpaceCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, uuid.UUID("059d78af-e675-498e-8b77-d4b2b4b9d4e7")).
			Return(&users.ExampleAlice, nil).Once()
		spacesMock.On("Create", mock.Anything, &spaces.CreateCmd{
			User:   &users.ExampleAlice,
			Name:   "Personal",
			Owners: []uuid.UUID{"059d78af-e675-498e-8b77-d4b2b4b9d4e7"},
		}).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		fsMock.On("CreateFS", mock.Anything, &users.ExampleAlice, &spaces.ExampleAlicePersonalSpace).
			Return(&dfs.ExampleAliceRoot, nil).Once()

		err := job.Run(ctx, json.RawMessage(`{"user-id": "059d78af-e675-498e-8b77-d4b2b4b9d4e7","name":"Personal","owners":["059d78af-e675-498e-8b77-d4b2b4b9d4e7"]}`))
		require.NoError(t, err)
	})

	t.Run("RunArgs with a GetByID error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewSpaceCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).
			Return(nil, errs.ErrNotFound).Once()

		err := job.RunArgs(ctx, &scheduler.SpaceCreateArgs{
			UserID: users.ExampleAlice.ID(),
			Name:   "Personal",
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		})
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("RunArgs with a spaces Create error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewSpaceCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).
			Return(&users.ExampleAlice, nil).Once()
		spacesMock.On("Create", mock.Anything, &spaces.CreateCmd{
			User:   &users.ExampleAlice,
			Name:   "Personal",
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		}).Return(nil, errs.ErrInternal).Once()

		err := job.RunArgs(ctx, &scheduler.SpaceCreateArgs{
			UserID: users.ExampleAlice.ID(),
			Name:   "Personal",
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		})
		require.ErrorIs(t, err, errs.ErrInternal)
	})

	t.Run("RunArgs with a CreateFS error", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		spacesMock := spaces.NewMockService(t)
		fsMock := dfs.NewMockService(t)
		job := NewSpaceCreateTaskRunner(usersMock, spacesMock, fsMock)

		usersMock.On("GetByID", mock.Anything, users.ExampleAlice.ID()).
			Return(&users.ExampleAlice, nil).Once()
		spacesMock.On("Create", mock.Anything, &spaces.CreateCmd{
			User:   &users.ExampleAlice,
			Name:   "Personal",
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		}).Return(&spaces.ExampleAlicePersonalSpace, nil).Once()
		fsMock.On("CreateFS", mock.Anything, &users.ExampleAlice, &spaces.ExampleAlicePersonalSpace).
			Return(nil, errs.ErrInternal).Once()

		err := job.RunArgs(ctx, &scheduler.SpaceCreateArgs{
			UserID: users.ExampleAlice.ID(),
			Name:   "Personal",
			Owners: []uuid.UUID{users.ExampleAlice.ID()},
		})
		require.ErrorIs(t, err, errs.ErrInternal)
	})
}
