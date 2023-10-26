package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSchedulerModels(t *testing.T) {
	t.Run("FileUploadArgs", func(t *testing.T) {
		err := FileUploadArgs{
			FolderID:   uuid.UUID("some-invalid-id"),
			Path:       "/foo/bar",
			FileID:     uuid.UUID("527c301f-2e66-4499-877e-0727a263268d"),
			UploadedAt: time.Now(),
		}.Validate()

		assert.EqualError(t, err, "folder-id: must be a valid UUID v4.")
	})

	t.Run("FSMoveArgs", func(t *testing.T) {
		err := FSMoveArgs{
			FolderID:    uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			SourceInode: uuid.UUID("some-invalid-id"),
			TargetPath:  "/foo/bar.txt",
			MovedAt:     time.Now(),
		}.Validate()

		assert.EqualError(t, err, "source-inode: must be a valid UUID v4.")
	})

	t.Run("UserCreateArgs", func(t *testing.T) {
		err := UserCreateArgs{
			UserID: uuid.UUID("some-invalid-id"),
		}.Validate()

		assert.EqualError(t, err, "user-id: must be a valid UUID v4.")
	})

	t.Run("FSGCArgs", func(t *testing.T) {
		err := FSGCArgs{}.Validate()

		assert.NoError(t, err)
	})
}
