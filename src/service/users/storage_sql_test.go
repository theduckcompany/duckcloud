package users

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StorageTestSuite struct {
	suite.Suite
	storage  *store
	nowData  time.Time
	userData User
}

func TestUserStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) SetupSuite() {
	t := suite.T()

	suite.nowData = time.Now().UTC()

	suite.userData = User{
		ID:        uuid.UUID("some-uuid"),
		Username:  "some-username",
		Email:     "some-email",
		password:  "some-password",
		CreatedAt: suite.nowData,
	}

	db, err := storage.NewSQliteDBWithMigrate(logger.NewNoop(), t.TempDir()+"/test.db", "../../../db/migration")
	require.NoError(t, err)

	suite.storage = newStorage(db)
}

func (suite *StorageTestSuite) Test_Create() {
	err := suite.storage.Save(context.Background(), &suite.userData)

	suite.Assert().NoError(err)
}

func (suite *StorageTestSuite) Test_GetByID() {
	res, err := suite.storage.GetByID(context.Background(), "some-uuid")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.NoError(err)
	suite.Equal(&suite.userData, res)
}

func (suite *StorageTestSuite) Test_GetByID_invalid_return_nil() {
	res, err := suite.storage.GetByID(context.Background(), "some-invalid-id")

	suite.NoError(err)
	suite.Nil(res)
}

func (suite *StorageTestSuite) Test_GetByEmail() {
	res, err := suite.storage.GetByEmail(context.Background(), "some-email")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.NoError(err)
	suite.Equal(&suite.userData, res)
}

func (suite *StorageTestSuite) Test_GetByEmail_invalid_return_nil() {
	res, err := suite.storage.GetByEmail(context.Background(), "some-invalid-email")

	suite.NoError(err)
	suite.Nil(res)
}

func (suite *StorageTestSuite) Test_GetByUsername() {
	res, err := suite.storage.GetByUsername(context.Background(), "some-username")

	suite.Require().NotNil(res)
	res.CreatedAt = res.CreatedAt.UTC()

	suite.Assert().NoError(err)
	suite.Assert().Equal(&suite.userData, res)
}

func (suite *StorageTestSuite) Test_GetByUsername_invalid_return_nil() {
	res, err := suite.storage.GetByUsername(context.Background(), "some-invalid-username")

	suite.NoError(err)
	suite.Nil(res)
}
