package taskstorage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	uuid "github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	now      = time.Now().UTC()
	nowAfter = time.Now().Add(time.Minute).UTC()
)

var ExampleFileUploadFileAlice = model.Task{
	ID:           uuid.UUID("c552f82a-fb8d-4655-8ccc-c0d0902d849d"),
	Priority:     1,
	Name:         "file-upload",
	Status:       model.Queuing,
	Retries:      0,
	RegisteredAt: now,
	Args:         json.RawMessage(`{"foo":"bar"}`),
}

var ExampleFileUploadFileAlice2 = model.Task{
	ID:           uuid.UUID("7ab0ae07-2383-4266-acfb-cdf0b0ec5312"),
	Priority:     2,
	Name:         "file-upload",
	Status:       model.Failed,
	Retries:      0,
	RegisteredAt: nowAfter,
	Args:         json.RawMessage(`{"foo":"bar"}`),
}

func Test_Tasks_SQLStorage(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	storage := NewSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := storage.Save(ctx, &ExampleFileUploadFileAlice)
		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(ctx, ExampleFileUploadFileAlice.ID)
		assert.Equal(t, &ExampleFileUploadFileAlice, res)
		require.NoError(t, err)
	})

	t.Run("GetNext success", func(t *testing.T) {
		task, err := storage.GetNext(ctx)
		require.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice, task)
	})

	t.Run("GetLastRegisteredTask success", func(t *testing.T) {
		res, err := storage.GetLastRegisteredTask(ctx, "file-upload")
		require.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice, res)
	})

	t.Run("Save success 2", func(t *testing.T) {
		// Save a second element in order to check that the
		// GetLastRegisteredTask get the taks in the correct order.
		err := storage.Save(ctx, &ExampleFileUploadFileAlice2)
		require.NoError(t, err)
	})

	t.Run("GetLastRegisteredTask success", func(t *testing.T) {
		res, err := storage.GetLastRegisteredTask(ctx, "file-upload")
		require.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice2, res)
	})

	t.Run("GetLastRegisteredTask with not tasks", func(t *testing.T) {
		res, err := storage.GetLastRegisteredTask(ctx, "unknown-task")
		assert.Nil(t, res)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := storage.Delete(ctx, ExampleFileUploadFileAlice.ID)
		require.NoError(t, err)

		res, err := storage.GetByID(ctx, ExampleFileUploadFileAlice.ID)
		assert.Nil(t, res)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Delete an already deleted task success", func(t *testing.T) {
		// Deleted by the previous test
		err := storage.Delete(ctx, ExampleFileUploadFileAlice.ID)
		require.NoError(t, err)
	})
}
