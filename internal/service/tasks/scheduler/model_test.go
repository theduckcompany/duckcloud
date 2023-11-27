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
			SpaceID:    uuid.UUID("some-invalid-id"),
			FileID:     uuid.UUID("527c301f-2e66-4499-877e-0727a263268d"),
			INodeID:    uuid.UUID("c87ebbda-435b-43b7-bab6-e93ca8f3831a"),
			UploadedAt: time.Now(),
		}.Validate()

		assert.EqualError(t, err, "space-id: must be a valid UUID v4.")
	})

	t.Run("FSMoveArgs", func(t *testing.T) {
		err := FSMoveArgs{
			SpaceID:     uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			SourceInode: uuid.UUID("some-invalid-id"),
			TargetPath:  "/foo/bar.txt",
			MovedAt:     time.Now(),
			MovedBy:     uuid.UUID("74926c6a-1802-45cd-bcb2-2dc0729fa986"),
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

	t.Run("FSRemoveDuplicateFileArgs", func(t *testing.T) {
		err := FSRemoveDuplicateFileArgs{
			ExistingFileID:  uuid.UUID("some-invalid-id"),
			DuplicateFileID: uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
		}.Validate()

		assert.EqualError(t, err, "existing-file-id: must be a valid UUID v4.")
	})

	t.Run("FSRefreshSizeArg", func(t *testing.T) {
		err := FSRefreshSizeArg{
			SpaceID:    uuid.UUID("a379fef3-ebc3-4069-b1ef-8c67948b3cff"),
			INodeID:    uuid.UUID("some-invalid-id"),
			ModifiedAt: time.Now().UTC(),
		}.Validate()

		assert.EqualError(t, err, "inode-id: must be a valid UUID v4.")
	})
}
