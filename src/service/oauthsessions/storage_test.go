package oauthsessions

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StorageTestSuite struct {
	suite.Suite
	storage     *sqlStorage
	nowData     time.Time
	sessionData Session
}

func TestSessionStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) SetupSuite() {
	t := suite.T()

	suite.nowData = time.Now().UTC()

	suite.sessionData = Session{
		AccessToken:      "some-access-token",
		AccessCreatedAt:  suite.nowData,
		AccessExpiresAt:  suite.nowData.Add(time.Hour),
		RefreshToken:     "some-refresh-token",
		RefreshCreatedAt: suite.nowData,
		RefreshExpiresAt: suite.nowData.Add(10 * time.Hour),
		ClientID:         "some-client-id",
		UserID:           "some-user-id",
		Scope:            "some-scope",
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, logger.NewNoop())
	require.NoError(t, err)

	suite.storage = newSqlStorage(db)
}

func (suite *StorageTestSuite) Test_Save() {
	err := suite.storage.Save(context.Background(), &suite.sessionData)

	suite.Assert().NoError(err)
}
