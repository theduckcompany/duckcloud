package dfs_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_DFS_Integration(t *testing.T) {
	ctx := context.Background()

	serv := startutils.NewServer(t)

	dfsSvc := serv.DFSSvc

	var space spaces.Space
	var spaceFS dfs.FS
	var rootFS *dfs.INode

	t.Run("CreateFS and RemoveFS success", func(t *testing.T) {
		fs := dfsSvc.GetSpaceFS(nil)
		tmpSpace, err := fs.CreateFS(ctx, &users.ExampleAlice, []uuid.UUID{serv.User.ID()})
		require.NoError(t, err)

		// Check that a new space have been created
		spaces, err := serv.SpacesSvc.GetAllUserSpaces(ctx, serv.User.ID(), nil)
		require.NoError(t, err)
		require.Len(t, spaces, 2) // the default one + the new one

		// Delete the new space
		err = fs.Destroy(ctx, tmpSpace)
		require.NoError(t, err)

		// Check that a new space have been created
		spaces, err = serv.SpacesSvc.GetAllUserSpaces(ctx, serv.User.ID(), nil)
		require.NoError(t, err)
		require.Len(t, spaces, 1) // only the default one
	})

	t.Run("Retrieve the default user space", func(t *testing.T) {
		spaces, err := serv.SpacesSvc.GetAllUserSpaces(ctx, serv.User.ID(), nil)
		require.NoError(t, err)
		require.Len(t, spaces, 1)

		space = spaces[0]
		spaceFS = dfsSvc.GetSpaceFS(&space)
	})

	t.Run("Get the rootFS success", func(t *testing.T) {
		var err error
		rootFS, err = spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/"})
		require.NoError(t, err)

		require.NotEmpty(t, rootFS)
		require.Nil(t, rootFS.Parent()) // The rootFS is the only inode without parent.

		require.Equal(t, "", rootFS.Name())
		require.True(t, rootFS.IsDir())
		require.WithinDuration(t, time.Now(), rootFS.LastModifiedAt(), time.Second)
	})

	t.Run("ListDir with an empty directory", func(t *testing.T) {
		dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/"}, nil)
		require.NoError(t, err)
		require.Len(t, dirContent, 0)
	})

	t.Run("ListDir with an unexisting path", func(t *testing.T) {
		dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/dir/doesn't/exists"}, nil)
		require.Nil(t, dirContent)
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	t.Run("CreateDir success", func(t *testing.T) {
		dir, err := spaceFS.CreateDir(ctx, &dfs.CreateDirCmd{
			Space:     &space,
			FilePath:  "/Documents/",
			CreatedBy: serv.User,
		})
		require.NoError(t, err)

		require.Equal(t, "Documents", dir.Name())
		require.Equal(t, uint64(0), dir.Size())
		require.Nil(t, dir.FileID())
		require.Equal(t, ptr.To(rootFS.ID()), dir.Parent()) // It have a parent and this is the root ("/")
		require.True(t, dir.IsDir())
		require.WithinDuration(t, time.Now(), dir.LastModifiedAt(), 30*time.Millisecond)

		// TODO: Check that the modified date have been modified for all the parents
		// newRootFS, err := spaceFS.Get(ctx, "/")
		// require.NoError(t, err)
		// require.NotEqual(t, newRootFS.LastModifiedAt())
	})

	t.Run("ListDir with 1 element inside the directory", func(t *testing.T) {
		dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/"}, nil)
		require.NoError(t, err)
		require.Len(t, dirContent, 1)

		require.Equal(t, "Documents", dirContent[0].Name())
	})

	t.Run("CreateDir with several inexisting level", func(t *testing.T) {
		var err error
		var dirBaz, dirBar, dirFoo *dfs.INode

		t.Run("Create the /foo/bar/baz directory", func(t *testing.T) {
			dirBaz, err = spaceFS.CreateDir(ctx, &dfs.CreateDirCmd{
				Space:     &space,
				FilePath:  "/foo/bar/baz",
				CreatedBy: serv.User,
			})
			require.NoError(t, err)

			require.Equal(t, "baz", dirBaz.Name())
			require.Equal(t, uint64(0), dirBaz.Size())
			require.Nil(t, dirBaz.FileID())
			require.True(t, dirBaz.IsDir())
			require.WithinDuration(t, time.Now(), dirBaz.LastModifiedAt(), 30*time.Millisecond)
		})

		t.Run("/foo/bar/bar have /foo/bar as parent", func(t *testing.T) {
			dirBar, err = spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/foo/bar"})
			require.NoError(t, err)
			assert.Equal(t, ptr.To(dirBar.ID()), dirBaz.Parent())
		})

		t.Run("/foo/bar have /foo as parent", func(t *testing.T) {
			dirFoo, err = spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/foo"})
			require.NoError(t, err)
			require.Equal(t, ptr.To(dirFoo.ID()), dirBar.Parent())
		})

		t.Run("/foo have / as parent", func(t *testing.T) {
			require.Equal(t, ptr.To(rootFS.ID()), dirFoo.Parent())
		})
	})

	t.Run("ListDir with 2 element and a pagination", func(t *testing.T) {
		t.Run("Get only the first element", func(t *testing.T) {
			dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/"}, &storage.PaginateCmd{Limit: 1})
			require.NoError(t, err)
			require.Len(t, dirContent, 1)
			require.Equal(t, "Documents", dirContent[0].Name())
		})

		t.Run("Get only the second element", func(t *testing.T) {
			dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/"}, &storage.PaginateCmd{
				Limit:      1,
				StartAfter: map[string]string{"name": "Documents"},
			})
			require.NoError(t, err)
			require.Len(t, dirContent, 1)
			require.Equal(t, "foo", dirContent[0].Name())
		})

		t.Run("Get both", func(t *testing.T) {
			dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/"}, &storage.PaginateCmd{Limit: 2})
			require.NoError(t, err)
			require.Len(t, dirContent, 2)
		})

		t.Run("Get two elements after Documents, it should return only one", func(t *testing.T) {
			dirContent, err := spaceFS.ListDir(ctx, &dfs.PathCmd{Space: &space, Path: "/"}, &storage.PaginateCmd{
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

			err := spaceFS.Upload(ctx, &dfs.UploadCmd{
				Space:      &space,
				FilePath:   "/Documents/todo.txt",
				Content:    buf,
				UploadedBy: serv.User,
			})
			require.NoError(t, err)
		})

		t.Run("Run the tasks", func(t *testing.T) {
			err := serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("Get the new file", func(t *testing.T) {
			info, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/Documents/todo.txt"})
			require.NoError(t, err)

			require.Equal(t, "todo.txt", info.Name())
			require.False(t, info.IsDir())
			require.Equal(t, uint64(len(content)), info.Size())
			require.WithinDuration(t, time.Now(), info.LastModifiedAt(), 30*time.Millisecond)
			modTime = info.LastModifiedAt()
		})

		t.Run("Download the new file", func(t *testing.T) {
			// Download the newly created file
			reader, err := spaceFS.Download(ctx, &dfs.PathCmd{Space: &space, Path: "/Documents/todo.txt"})
			require.NoError(t, err)

			res, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, []byte(content), res)
		})

		t.Run("Check parents size and modtime", func(t *testing.T) {
			root, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/"})
			require.NoError(t, err)
			assert.Equal(t, modTime, root.LastModifiedAt())
			assert.Equal(t, uint64(len(content)), root.Size())

			dir, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/Documents"})
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

			oldFile, err = spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/Documents/todo.txt"})
			require.NoError(t, err)
		})

		t.Run("Move", func(t *testing.T) {
			// The /NewSpace doesn't exists yet. It must be automatically created
			err := spaceFS.Move(ctx, &dfs.MoveCmd{
				Src:     &dfs.PathCmd{Space: &space, Path: "/Documents/todo.txt"},
				Dst:     &dfs.PathCmd{Space: &space, Path: "/NewDocuments/todo.txt"},
				MovedBy: serv.User,
			})
			require.NoError(t, err)
		})

		t.Run("Run the tasks", func(t *testing.T) {
			err := serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("Get the new file", func(t *testing.T) {
			var err error

			newFile, err = spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/NewDocuments/todo.txt"})
			require.NoError(t, err)

			// A move must change the inode path an keep the same id
			require.Equal(t, oldFile.ID(), newFile.ID())
			require.Equal(t, "todo.txt", newFile.Name())
			require.NotEqual(t, oldFile.LastModifiedAt(), newFile.LastModifiedAt())
			require.WithinDuration(t, time.Now(), newFile.LastModifiedAt(), 30*time.Millisecond)
		})

		t.Run("Check old parents modtime and size", func(t *testing.T) {
			dir, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/Documents"})
			require.NoError(t, err)
			assert.Equal(t, newFile.LastModifiedAt(), dir.LastModifiedAt())
			// Theres is no more files so the size is 0
			assert.Equal(t, uint64(0), dir.Size())
		})

		t.Run("Check new parents modtime and size", func(t *testing.T) {
			root, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/"})
			require.NoError(t, err)
			assert.Equal(t, newFile.LastModifiedAt(), root.LastModifiedAt())
			assert.Equal(t, newFile.Size(), root.Size())

			dir, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/NewDocuments"})
			require.NoError(t, err)
			assert.Equal(t, newFile.LastModifiedAt(), dir.LastModifiedAt())
			assert.Equal(t, newFile.Size(), dir.Size())
		})
	})

	t.Run("Check duplicate files", func(t *testing.T) {
		content := "Hello, World!"

		t.Run("Create the test directory", func(t *testing.T) {
			_, err := spaceFS.CreateDir(ctx, &dfs.CreateDirCmd{
				Space:     &space,
				FilePath:  "/Duplicate",
				CreatedBy: serv.User,
			})
			require.NoError(t, err)
		})

		t.Run("Upload the first file", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			buf.WriteString(content)

			err := spaceFS.Upload(ctx, &dfs.UploadCmd{
				Space:      &space,
				FilePath:   "/Duplicate/todo.txt",
				Content:    buf,
				UploadedBy: serv.User,
			})
			require.NoError(t, err)

			err = serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("Upload the second same file", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			buf.WriteString(content)

			err := spaceFS.Upload(ctx, &dfs.UploadCmd{
				Space:      &space,
				FilePath:   "/Duplicate/todo-duplicate.txt",
				Content:    buf,
				UploadedBy: serv.User,
			})
			require.NoError(t, err)

			err = serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("The two file must have the same fileID", func(t *testing.T) {
			file1, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/Duplicate/todo.txt"})
			require.NoError(t, err)

			file2, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/Duplicate/todo-duplicate.txt"})
			require.NoError(t, err)

			require.Equal(t, file1.FileID(), file2.FileID())
		})

		t.Run("The first replicate is deleted, the second still have the file", func(t *testing.T) {
			err := spaceFS.Remove(ctx, &dfs.PathCmd{Space: &space, Path: "/Duplicate/todo.txt"})
			require.NoError(t, err)

			err = serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)

			reader, err := spaceFS.Download(ctx, &dfs.PathCmd{Space: &space, Path: "/Duplicate/todo-duplicate.txt"})
			require.NoError(t, err)

			res, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, []byte(content), res)
		})

		t.Run("Delete the directory", func(t *testing.T) {
			err := spaceFS.Remove(ctx, &dfs.PathCmd{Space: &space, Path: "/Duplicate"})
			require.NoError(t, err)

			err = serv.RunnerSvc.Run(ctx)
			require.NoError(t, err)
		})

		t.Run("Move with the same name", func(t *testing.T) {
			t.Run("Setup", func(t *testing.T) {
				_, err := spaceFS.CreateDir(ctx, &dfs.CreateDirCmd{
					Space:     &space,
					FilePath:  "/move-same-name",
					CreatedBy: serv.User,
				})
				require.NoError(t, err)

				err = spaceFS.Upload(ctx, &dfs.UploadCmd{
					Space:      &space,
					FilePath:   "/move-same-name/foo.txt",
					Content:    http.NoBody,
					UploadedBy: serv.User,
				})
				require.NoError(t, err)

				err = serv.RunnerSvc.Run(ctx)
				require.NoError(t, err)
			})

			t.Run("Move success", func(t *testing.T) {
				file, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/move-same-name/foo.txt"})
				require.NoError(t, err)

				res, err := spaceFS.Rename(ctx, file, "foo2.txt")
				require.NoError(t, err)

				require.Equal(t, "foo2.txt", res.Name())
			})

			t.Run("Move with a file with the same name", func(t *testing.T) {
				err := spaceFS.Upload(ctx, &dfs.UploadCmd{
					Space:      &space,
					FilePath:   "/move-same-name/bar.txt",
					Content:    http.NoBody,
					UploadedBy: serv.User,
				})
				require.NoError(t, err)

				err = serv.RunnerSvc.Run(ctx)
				require.NoError(t, err)

				file, err := spaceFS.Get(ctx, &dfs.PathCmd{Space: &space, Path: "/move-same-name/foo2.txt"})
				require.NoError(t, err)

				res, err := spaceFS.Rename(ctx, file, "bar.txt")
				require.NoError(t, err)

				require.Equal(t, "bar (1).txt", res.Name())
			})

			t.Run("Destroy", func(t *testing.T) {
				err := spaceFS.Remove(ctx, &dfs.PathCmd{Space: &space, Path: "/move-same-name"})
				require.NoError(t, err)

				err = serv.RunnerSvc.Run(ctx)
				require.NoError(t, err)
			})
		})
	})
}
