package dfs_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_DFS_Integration(t *testing.T) {
	ctx := context.Background()

	serv := startutils.NewServer(t)
	serv.Bootstrap(t)

	dfsSvc := serv.DFSSvc

	var folder folders.Folder
	var folderFS dfs.FS
	var rootFS *dfs.INode

	t.Run("CreateFS and RemoveFS success", func(t *testing.T) {
		tmpFolder, err := dfsSvc.CreateFS(ctx, []uuid.UUID{serv.User.ID()})
		require.NoError(t, err)

		// Check that a new folder have been created
		folders, err := serv.FoldersSvc.GetAllUserFolders(ctx, serv.User.ID(), nil)
		require.NoError(t, err)
		require.Len(t, folders, 2) // the default one + the new one

		// Delete the new folder
		err = dfsSvc.RemoveFS(ctx, tmpFolder)
		require.NoError(t, err)

		// Check that a new folder have been created
		folders, err = serv.FoldersSvc.GetAllUserFolders(ctx, serv.User.ID(), nil)
		require.NoError(t, err)
		require.Len(t, folders, 1) // only the default one
	})

	t.Run("Retrieve the default user folder", func(t *testing.T) {
		folders, err := serv.FoldersSvc.GetAllUserFolders(ctx, serv.User.ID(), nil)
		require.NoError(t, err)
		require.Len(t, folders, 1)

		folder = folders[0]
		folderFS = dfsSvc.GetFolderFS(&folder)
	})

	t.Run("Get the rootFS success", func(t *testing.T) {
		var err error
		rootFS, err = folderFS.Get(ctx, "/")
		require.NoError(t, err)

		require.NotEmpty(t, rootFS)
		require.Nil(t, rootFS.Parent()) // The rootFS is the only inode without parent.

		require.Equal(t, "", rootFS.Name())
		require.True(t, rootFS.IsDir())
		require.Equal(t, uint64(0), rootFS.Size())
		require.Empty(t, rootFS.Checksum())
		require.WithinDuration(t, time.Now(), rootFS.LastModifiedAt(), 14*time.Millisecond)
	})

	t.Run("ListDir with an empty directory", func(t *testing.T) {
		dirContent, err := folderFS.ListDir(ctx, "/", nil)
		require.NoError(t, err)
		require.Len(t, dirContent, 0)
	})

	t.Run("ListDir with an unexisting path", func(t *testing.T) {
		dirContent, err := folderFS.ListDir(ctx, "/dir/doesn't/exists", nil)
		require.Nil(t, dirContent)
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		dir, err := folderFS.CreateDir(ctx, "/Documents/")
		require.NoError(t, err)

		require.Equal(t, "Documents", dir.Name())
		require.Equal(t, uint64(0), dir.Size())
		require.Nil(t, dir.FileID())
		require.Equal(t, ptr.To(rootFS.ID()), dir.Parent()) // It have a parent and this is the root ("/")
		require.True(t, dir.IsDir())
		require.Empty(t, dir.Checksum())
		require.WithinDuration(t, time.Now(), dir.LastModifiedAt(), 20*time.Millisecond)

		// TODO: Check that the modified date have been modified for all the parents
		// newRootFS, err := folderFS.Get(ctx, "/")
		// require.NoError(t, err)
		// require.NotEqual(t, newRootFS.LastModifiedAt())
	})

	t.Run("ListDir with 1 element inside the directory", func(t *testing.T) {
		dirContent, err := folderFS.ListDir(ctx, "/", nil)
		require.NoError(t, err)
		require.Len(t, dirContent, 1)

		require.Equal(t, "Documents", dirContent[0].Name())
	})

	t.Run("CreateDir with several inexisting level", func(t *testing.T) {
		var err error
		var dirBaz, dirBar, dirFoo *dfs.INode

		t.Run("Create the /foo/bar/baz directory", func(t *testing.T) {
			dirBaz, err = folderFS.CreateDir(ctx, "/foo/bar/baz")
			require.NoError(t, err)

			require.Equal(t, "baz", dirBaz.Name())
			require.Equal(t, uint64(0), dirBaz.Size())
			require.Nil(t, dirBaz.FileID())
			require.True(t, dirBaz.IsDir())
			require.Empty(t, dirBaz.Checksum())
			require.WithinDuration(t, time.Now(), dirBaz.LastModifiedAt(), 20*time.Millisecond)
		})

		t.Run("/foo/bar/bar have /foo/bar as parent", func(t *testing.T) {
			dirBar, err = folderFS.Get(ctx, "/foo/bar")
			require.NoError(t, err)
			assert.Equal(t, ptr.To(dirBar.ID()), dirBaz.Parent())
		})

		t.Run("/foo/bar have /foo as parent", func(t *testing.T) {
			dirFoo, err = folderFS.Get(ctx, "/foo")
			require.NoError(t, err)
			require.Equal(t, ptr.To(dirFoo.ID()), dirBar.Parent())
		})

		t.Run("/foo have / as parent", func(t *testing.T) {
			require.Equal(t, ptr.To(rootFS.ID()), dirFoo.Parent())
		})
	})

	t.Run("ListDir with 2 element and a pagination", func(t *testing.T) {
		t.Run("Get only the first element", func(t *testing.T) {
			dirContent, err := folderFS.ListDir(ctx, "/", &storage.PaginateCmd{Limit: 1})
			require.NoError(t, err)
			require.Len(t, dirContent, 1)
			require.Equal(t, "Documents", dirContent[0].Name())
		})

		t.Run("Get only the second element", func(t *testing.T) {
			dirContent, err := folderFS.ListDir(ctx, "/", &storage.PaginateCmd{
				Limit:      1,
				StartAfter: map[string]string{"name": "Documents"},
			})
			require.NoError(t, err)
			require.Len(t, dirContent, 1)
			require.Equal(t, "foo", dirContent[0].Name())
		})

		t.Run("Get both", func(t *testing.T) {
			dirContent, err := folderFS.ListDir(ctx, "/", &storage.PaginateCmd{Limit: 2})
			require.NoError(t, err)
			require.Len(t, dirContent, 2)
		})

		t.Run("Get two elements after Documents, it should return only one", func(t *testing.T) {
			dirContent, err := folderFS.ListDir(ctx, "/", &storage.PaginateCmd{
				Limit:      2,
				StartAfter: map[string]string{"name": "Documents"},
			})
			require.NoError(t, err)
			require.Len(t, dirContent, 1)
			require.Equal(t, "foo", dirContent[0].Name())
		})
	})

	t.Run("Upload and Download success", func(t *testing.T) {
		content := "Hello, World!"
		var modTime time.Time

		t.Run("Upload the file", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			buf.WriteString(content)

			err := folderFS.Upload(ctx, "/Documents/todo.txt", buf)
			require.NoError(t, err)
		})

		t.Run("Run the tasks", func(t *testing.T) {
			err := serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("Get the new file", func(t *testing.T) {
			info, err := folderFS.Get(ctx, "/Documents/todo.txt")
			require.NoError(t, err)

			require.Equal(t, "todo.txt", info.Name())
			require.False(t, info.IsDir())
			require.Equal(t, uint64(len(content)), info.Size())
			require.Equal(t, "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=", info.Checksum())
			require.WithinDuration(t, time.Now(), info.LastModifiedAt(), 20*time.Millisecond)
			modTime = info.LastModifiedAt()
		})

		t.Run("Download the new file", func(t *testing.T) {
			// Download the newly created file
			reader, err := folderFS.Download(ctx, "/Documents/todo.txt")
			require.NoError(t, err)

			res, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, []byte(content), res)
		})

		t.Run("Check parents size and modtime", func(t *testing.T) {
			root, err := folderFS.Get(ctx, "/")
			require.NoError(t, err)
			assert.Equal(t, modTime, root.LastModifiedAt())
			assert.Equal(t, uint64(len(content)), root.Size())

			dir, err := folderFS.Get(ctx, "/Documents")
			require.NoError(t, err)
			assert.Equal(t, modTime, dir.LastModifiedAt())
			assert.Equal(t, uint64(len(content)), dir.Size())
		})
	})

	t.Run("Move success", func(t *testing.T) {
		var oldFile *dfs.INode
		var newFile *dfs.INode

		t.Run("Get the old file", func(t *testing.T) {
			var err error

			oldFile, err = folderFS.Get(ctx, "/Documents/todo.txt")
			require.NoError(t, err)
		})

		t.Run("Move", func(t *testing.T) {
			// The /NewFolder doesn't exists yet. It must be automatically created
			err := folderFS.Move(ctx, "/Documents/todo.txt", "/NewDocuments/todo.txt")
			require.NoError(t, err)
		})

		t.Run("Run the tasks", func(t *testing.T) {
			err := serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("Get the new file", func(t *testing.T) {
			var err error

			newFile, err = folderFS.Get(ctx, "/NewDocuments/todo.txt")
			require.NoError(t, err)

			// A move must change the inode path an keep the same id
			require.Equal(t, oldFile.ID(), newFile.ID())
			require.Equal(t, "todo.txt", newFile.Name())
			require.NotEqual(t, oldFile.LastModifiedAt(), newFile.LastModifiedAt())
			require.WithinDuration(t, time.Now(), newFile.LastModifiedAt(), 20*time.Millisecond)
		})

		t.Run("Check old parents modtime and size", func(t *testing.T) {
			dir, err := folderFS.Get(ctx, "/Documents")
			require.NoError(t, err)
			assert.Equal(t, newFile.LastModifiedAt(), dir.LastModifiedAt())
			// Theres is no more files so the size is 0
			assert.Equal(t, uint64(0), dir.Size())
		})

		t.Run("Check new parents modtime and size", func(t *testing.T) {
			root, err := folderFS.Get(ctx, "/")
			require.NoError(t, err)
			assert.Equal(t, newFile.LastModifiedAt().Add(time.Microsecond), root.LastModifiedAt())
			assert.Equal(t, newFile.Size(), root.Size())

			dir, err := folderFS.Get(ctx, "/NewDocuments")
			require.NoError(t, err)
			assert.Equal(t, newFile.LastModifiedAt().Add(time.Microsecond), dir.LastModifiedAt())
			assert.Equal(t, newFile.Size(), dir.Size())
		})
	})
}
