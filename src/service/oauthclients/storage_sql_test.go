package oauthclients

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StorageTestSuite struct {
	suite.Suite
	storage    *sqlStorage
	nowData    time.Time
	clientData Client
}

func TestClientStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) SetupSuite() {
	t := suite.T()

	suite.nowData = time.Now().UTC()

	userID := "some-userID"
	suite.clientData = Client{
		ID:             uuid.UUID("some-uuid"),
		Secret:         "some-secret",
		RedirectURI:    "some-url",
		UserID:         &userID,
		CreatedAt:      suite.nowData,
		Scopes:         []string{"scope-a"},
		IsPublic:       true,
		SkipValidation: true,
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, logger.NewNoop())
	require.NoError(t, err)

	suite.storage = newSqlStorage(db)
}

func (suite *StorageTestSuite) Test_Create() {
	err := suite.storage.Save(context.Background(), &suite.clientData)

	suite.Assert().NoError(err)
}

func (suite *StorageTestSuite) Test_GetByID() {
	res, err := suite.storage.GetByID(context.Background(), uuid.UUID("some-uuid"))

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.NoError(err)
	suite.Equal(&suite.clientData, res)
}

func (suite *StorageTestSuite) Test_GetByEmail_invalid_return_nil() {
	res, err := suite.storage.GetByID(context.Background(), "some-invalid-uuid")

	suite.NoError(err)
	suite.Nil(res)
}
