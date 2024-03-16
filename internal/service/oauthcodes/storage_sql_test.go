package oauthcodes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestOauthCodeSQLStorage(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	storage := newSqlStorage(db)

	user := users.NewFakeUser(t).BuildAndStore(ctx, db)
	oauthclients := oauthclients.NewFakeClient(t).CreatedBy(user).BuildAndStore(ctx, db)
	code := NewFakeCode(t).WithClient(oauthclients).CreatedBy(user).Build()

	t.Run("save", func(t *testing.T) {
		// Run
		err := storage.Save(ctx, code)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID ok", func(t *testing.T) {
		// Run
		res, err := storage.GetByCode(context.Background(), code.code)

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, code, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		code, err := storage.GetByCode(ctx, secret.NewText("some-invalid-code"))

		// Asserts
		assert.Nil(t, code)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByCode success", func(t *testing.T) {
		// Run
		err := storage.RemoveByCode(ctx, code.code)

		// Asserts
		require.NoError(t, err)
		// Check that the code is no more available
		code, err := storage.GetByCode(ctx, code.code)
		assert.Nil(t, code)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByCode invalid code", func(t *testing.T) {
		// Run
		err := storage.RemoveByCode(ctx, secret.NewText("some-invalid-code"))

		// Asserts
		require.NoError(t, err)
	})
}
