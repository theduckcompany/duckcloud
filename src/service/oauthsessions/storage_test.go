package oauthsessions

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionStorageStorage(t *testing.T) {
	tools := tools.NewMock(t)

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, tools)
	require.NoError(t, err)

	nowData := time.Now().UTC()

	sessionData := Session{
		AccessToken:      "some-access-token",
		AccessCreatedAt:  nowData,
		AccessExpiresAt:  nowData.Add(time.Hour),
		RefreshToken:     "some-refresh-token",
		RefreshCreatedAt: nowData,
		RefreshExpiresAt: nowData.Add(10 * time.Hour),
		ClientID:         "some-client-id",
		UserID:           "some-user-id",
		Scope:            "some-scope",
	}

	storage := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := storage.Save(context.Background(), &sessionData)
		assert.NoError(t, err)
	})
}
