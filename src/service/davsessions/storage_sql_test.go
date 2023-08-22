package davsessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestDavSessionSqlStorage(t *testing.T) {
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(context.Background(), &ExampleAliceSession)

		assert.NoError(t, err)
	})

	t.Run("GetByUsernameAndPassHash success", func(t *testing.T) {
		res, err := store.GetByUsernameAndPassHash(context.Background(), ExampleAliceSession.username, ExampleAliceSession.password)

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("GetByUsernameAndPassHash not found", func(t *testing.T) {
		res, err := store.GetByUsernameAndPassHash(context.Background(), "some-invalid-username", "some-hashed-password")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		res, err := store.GetAllForUser(context.Background(), ExampleAliceSession.userID, &storage.PaginateCmd{Limit: 10})

		assert.NoError(t, err)
		assert.Equal(t, []DavSession{ExampleAliceSession}, res)
	})

	t.Run("GetAllForUser not found", func(t *testing.T) {
		res, err := store.GetAllForUser(context.Background(), uuid.UUID("unknown-id"), &storage.PaginateCmd{Limit: 10})

		assert.NoError(t, err)
		assert.Equal(t, []DavSession{}, res)
	})
}
