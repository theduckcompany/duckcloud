package uploads

import (
	context "context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestINodeSqlstore(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAliceUpload)
		assert.NoError(t, err)
	})

	t.Run("GetAll success", func(t *testing.T) {
		res, err := store.GetAll(ctx, &storage.PaginateCmd{Limit: 2})
		assert.NoError(t, err)

		assert.Equal(t, []Upload{ExampleAliceUpload}, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := store.Delete(ctx, ExampleAliceUpload.id)
		assert.NoError(t, err)
	})

	t.Run("Delete an unknown id", func(t *testing.T) {
		err := store.Delete(ctx, uuid.UUID("some-unknown-id"))
		assert.NoError(t, err)
	})
}
