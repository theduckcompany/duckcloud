package server

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestBootstrap(t *testing.T) {
	ctx := context.Background()
	tools := tools.NewMock(t)
	fs := afero.NewMemMapFs()

	cfg := NewDefaultConfig()
	db := storage.NewTestStorage(t)

	err := storage.RunMigrations(cfg.Storage, db, tools)
	require.NoError(t, err)

	user := users.CreateCmd{
		Username: "foo",
		Password: "qwert1234",
		IsAdmin:  true,
	}

	err = Bootstrap(ctx, db, fs, cfg, user)
	require.NoError(t, err)

	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db, inodesSvc)
	usersSvc := users.Init(tools, db, inodesSvc, foldersSvc)

	tools.PasswordMock.On("Compare", mock.Anything, mock.AnythingOfType("string"), "qwert1234").Return(nil).Once()

	res, err := usersSvc.Authenticate(ctx, user.Username, user.Password)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
