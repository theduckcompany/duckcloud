package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
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

	db := storage.NewTestStorage(t)
	storage := NewSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := storage.Save(ctx, &ExampleFileUploadFileAlice)
		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(ctx, ExampleFileUploadFileAlice.ID)
		assert.Equal(t, &ExampleFileUploadFileAlice, res)
		assert.NoError(t, err)
	})

	t.Run("GetNext success", func(t *testing.T) {
		task, err := storage.GetNext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice, task)
	})

	t.Run("GetLastRegisteredTask success", func(t *testing.T) {
		res, err := storage.GetLastRegisteredTask(ctx, "file-upload")
		assert.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice, res)
	})

	t.Run("Save success 2", func(t *testing.T) {
		// Save a second element in order to check that the
		// GetLastRegisteredTask get the taks in the correct order.
		err := storage.Save(ctx, &ExampleFileUploadFileAlice2)
		assert.NoError(t, err)
	})

	t.Run("GetLastRegisteredTask success", func(t *testing.T) {
		res, err := storage.GetLastRegisteredTask(ctx, "file-upload")
		assert.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice2, res)
	})

	t.Run("GetLastRegisteredTask with not tasks", func(t *testing.T) {
		res, err := storage.GetLastRegisteredTask(ctx, "unknown-task")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := storage.Delete(ctx, ExampleFileUploadFileAlice.ID)
		assert.NoError(t, err)

		res, err := storage.GetByID(ctx, ExampleFileUploadFileAlice.ID)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Delete an already deleted task success", func(t *testing.T) {
		// Deleted by the previous test
		err := storage.Delete(ctx, ExampleFileUploadFileAlice.ID)
		assert.NoError(t, err)
	})
}
