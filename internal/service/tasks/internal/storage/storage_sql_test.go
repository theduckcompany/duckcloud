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

var now = time.Now().UTC()

var ExampleFileUploadFileAlice = model.Task{
	ID:           uuid.UUID("c552f82a-fb8d-4655-8ccc-c0d0902d849d"),
	Priority:     1,
	Name:         "file-upload",
	Status:       model.Queuing,
	Retries:      0,
	RegisteredAt: now,
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

	t.Run("GetNext success", func(t *testing.T) {
		task, err := storage.GetNext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleFileUploadFileAlice, task)
	})
}
