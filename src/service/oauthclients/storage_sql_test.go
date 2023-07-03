package oauthclients

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
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
	tools := tools.NewMock(t)

	suite.nowData = time.Now().UTC()

	userID := "some-userID"
	suite.clientData = Client{
		ID:             "some-client-id",
		Secret:         "some-secret",
		RedirectURI:    "some-url",
		UserID:         &userID,
		CreatedAt:      suite.nowData,
		Scopes:         []string{"scope-a"},
		Public:         true,
		SkipValidation: true,
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, tools)
	require.NoError(t, err)

	suite.storage = newSqlStorage(db)
}

func (suite *StorageTestSuite) Test_Create() {
	err := suite.storage.Save(context.Background(), &suite.clientData)

	suite.Assert().NoError(err)
}

func (suite *StorageTestSuite) Test_GetByID() {
	res, err := suite.storage.GetByID(context.Background(), "some-client-id")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.NoError(err)
	suite.Equal(&suite.clientData, res)
}

func (suite *StorageTestSuite) Test_GetByEmail_invalid_return_nil() {
	res, err := suite.storage.GetByID(context.Background(), "some-inval##id-uuid")

	suite.NoError(err)
	suite.Nil(res)
}
