package server

import (
	"context"
	"testing"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBootstrap(t *testing.T) {
	ctx := context.Background()
	tools := tools.NewMock(t)

	cfg := NewDefaultConfig()
	cfg.Storage.Path = t.TempDir() + "/db.sqlite"
	err := storage.RunMigrations(cfg.Storage, tools)
	require.NoError(t, err)

	user := users.CreateCmd{
		Username: "foo",
		Email:    "foo@bar.baz",
		Password: "qwert1234",
	}

	err = Bootstrap(ctx, cfg, user)
	require.NoError(t, err)

	db, err := storage.NewSQliteClient(cfg.Storage, tools.Logger())
	require.NoError(t, err)
	inodesSvc := inodes.Init(tools, db)
	usersSvc := users.Init(tools, db, inodesSvc)

	tools.PasswordMock.On("Compare", mock.Anything, mock.AnythingOfType("string"), "qwert1234").Return(nil).Once()

	res, err := usersSvc.Authenticate(ctx, user.Username, user.Password)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
