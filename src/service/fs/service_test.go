package fs

import (
	"context"
	"os"
	"testing"
	"testing/fstest"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/logger"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_FS(t *testing.T) {
	ctx := context.Background()
	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)

	inodesSvc := inodes.Init(tools, db)
	afs := afero.NewMemMapFs()
	filesSvc, err := files.NewFSService(afs, "/", tools.Logger())
	require.NoError(t, err)

	userID := uuid.UUID("fd801c11-356a-4abb-8d72-1ea87d2d7201")

	rootInode, err := inodesSvc.BootstrapUser(ctx, userID)
	require.NoError(t, err)

	duckFS := NewFSService(userID, rootInode.ID(), inodesSvc, filesSvc)

	err = duckFS.CreateDir(ctx, "foo", 0o700)
	require.NoError(t, err)

	file, err := duckFS.OpenFile(ctx, "foo/bar.txt", os.O_CREATE|os.O_RDWR, 0o700)
	require.NoError(t, err)

	ret, err := file.Write([]byte("Hello, World!"))
	require.NoError(t, err)
	require.Equal(t, ret, 13)

	assert.NoError(t, fstest.TestFS(duckFS, "foo/bar.txt"))
}
