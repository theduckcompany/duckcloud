package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
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
