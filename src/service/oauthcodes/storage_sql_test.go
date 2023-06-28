package oauthcodes

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
	storage  *sqlStorage
	nowData  time.Time
	codeData Code
}

func TestCodeStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) SetupSuite() {
	t := suite.T()

	suite.nowData = time.Now().UTC()

	suite.codeData = Code{
		Code:        "some-code",
		CreatedAt:   suite.nowData,
		ExpiresAt:   suite.nowData.Add(time.Hour),
		ClientID:    "some-client-id",
		UserID:      "some-user-id",
		RedirectURI: "http://some-redirect.com/uri",
		Scope:       "some-scope",
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, logger.NewNoop())
	require.NoError(t, err)

	suite.storage = newSqlStorage(db)
}

func (suite *StorageTestSuite) Test_Save() {
	err := suite.storage.Save(context.Background(), &suite.codeData)

	suite.Assert().NoError(err)
}
