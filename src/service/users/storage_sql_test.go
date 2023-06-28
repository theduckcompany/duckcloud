package users

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

type SqlStorageTestSuite struct {
	suite.Suite
	storage  *sqlStorage
	nowData  time.Time
	userData User
}

func TestUserSqlStorageSuite(t *testing.T) {
	suite.Run(t, new(SqlStorageTestSuite))
}

func (suite *SqlStorageTestSuite) SetupSuite() {
	t := suite.T()

	suite.nowData = time.Now().UTC()

	suite.userData = User{
		ID:        uuid.UUID("some-uuid"),
		Username:  "some-username",
		Email:     "some-email",
		password:  "some-password",
		CreatedAt: suite.nowData,
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, logger.NewNoop())
	require.NoError(t, err)

	suite.storage = newSqlStorage(db)
}

func (suite *SqlStorageTestSuite) Test_Create() {
	err := suite.storage.Save(context.Background(), &suite.userData)

	suite.Assert().NoError(err)
}

func (suite *SqlStorageTestSuite) Test_GetByID() {
	res, err := suite.storage.GetByID(context.Background(), "some-uuid")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.NoError(err)
	suite.Equal(&suite.userData, res)
}

func (suite *SqlStorageTestSuite) Test_GetByID_invalid_return_nil() {
	res, err := suite.storage.GetByID(context.Background(), "some-invalid-id")

	suite.NoError(err)
	suite.Nil(res)
}

func (suite *SqlStorageTestSuite) Test_GetByEmail() {
	res, err := suite.storage.GetByEmail(context.Background(), "some-email")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.NoError(err)
	suite.Equal(&suite.userData, res)
}

func (suite *SqlStorageTestSuite) Test_GetByEmail_invalid_return_nil() {
	res, err := suite.storage.GetByEmail(context.Background(), "some-invalid-email")

	suite.NoError(err)
	suite.Nil(res)
}

func (suite *SqlStorageTestSuite) Test_GetByUsername() {
	res, err := suite.storage.GetByUsername(context.Background(), "some-username")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.Assert().NoError(err)
	suite.Assert().Equal(&suite.userData, res)
}

func (suite *SqlStorageTestSuite) Test_GetByUsername_invalid_return_nil() {
	res, err := suite.storage.GetByUsername(context.Background(), "some-invalid-username")

	suite.NoError(err)
	suite.Nil(res)
}
