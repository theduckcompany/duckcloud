package davsessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDavSessionSqlStorage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db)

	// Data
	user := users.NewFakeUser(t).WithAdminRole().BuildAndStore(ctx, db)
	space := spaces.NewFakeSpace(t).WithOwners(*user).BuildAndStore(ctx, db)

	sessionPassword := "some-password"
	session := NewFakeSession(t).
		WithPassword(sessionPassword).
		CreatedBy(user).
		WithSpace(space).
		Build()

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := store.Save(context.Background(), session)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByUsernameAndPassword success", func(t *testing.T) {
		// Run
		res, err := store.GetByUsernameAndPassword(context.Background(), session.username, session.password)

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, session, res)
	})

	t.Run("GetByUsernameAndPassword not found", func(t *testing.T) {
		// Run
		res, err := store.GetByUsernameAndPassword(context.Background(), "some-invalid-username", secret.NewText("some-hashed-password"))

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetByID success", func(t *testing.T) {
		// Run
		res, err := store.GetByID(context.Background(), session.id)

		// Asserts
		assert.Equal(t, session, res)
		require.NoError(t, err)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		res, err := store.GetByID(context.Background(), "some-invalid-id")

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		// Run
		res, err := store.GetAllForUser(context.Background(), session.userID, &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []DavSession{*session}, res)
	})

	t.Run("GetAllForUser not found", func(t *testing.T) {
		// Run
		res, err := store.GetAllForUser(context.Background(), uuid.UUID("unknown-id"), &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []DavSession{}, res)
	})

	t.Run("RemoveByID success", func(t *testing.T) {
		// Run
		err := store.RemoveByID(context.Background(), session.ID())
		require.NoError(t, err)

		// Asserts
		res, err := store.GetByID(context.Background(), session.id)
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})
}
