package dav

import (
	"context"
	stdfs "io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

func Test_DAVFS(t *testing.T) {
	t.Run("Stat success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.Stat(ctx, "foo/bar")
		assert.NoError(t, err)
		assert.EqualValues(t, &inodes.ExampleAliceFile, res)
	})

	t.Run("Stat with an invalid path", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		res, err := dav.Stat(ctx, "/foo/bar")
		assert.Nil(t, res)
		assert.EqualError(t, err, "stat /foo/bar: invalid argument")
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("RemoveAll", mock.Anything, "foo/bar").Return(nil).Once()

		err := dav.RemoveAll(ctx, "foo/bar")
		assert.NoError(t, err)
	})

	t.Run("RemoveAll with an invalid path", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		err := dav.RemoveAll(ctx, "/foo/bar")
		assert.EqualError(t, err, "removeAll /foo/bar: invalid argument")
	})

	t.Run("Mkdir success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("CreateDir", mock.Anything, "foo").Return(&inodes.ExampleAliceRoot, nil).Once()

		err := dav.Mkdir(ctx, "foo", 0o644)
		assert.NoError(t, err)
	})

	t.Run("OpenFile a folder success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo").Return(&inodes.ExampleAliceRoot, nil).Once()

		res, err := dav.OpenFile(ctx, "foo", os.O_RDONLY, 0o644)
		assert.NoError(t, err)

		// Check the file infos
		stats, err := res.Stat()
		require.NoError(t, err)
		assert.True(t, stats.IsDir())
		assert.Equal(t, inodes.ExampleAliceRoot.LastModifiedAt(), stats.ModTime())
		assert.Equal(t, int64(0), stats.Size())
		assert.Nil(t, stats.Sys())
		assert.Equal(t, stdfs.FileMode(0o660), stats.Mode().Perm())
		assert.True(t, stats.Mode().IsDir())
	})

	t.Run("OpenFile a file success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_RDONLY, 0o644)
		assert.NoError(t, err)

		// Check the file infos
		stats, err := res.Stat()
		require.NoError(t, err)
		assert.False(t, stats.IsDir())
		assert.Equal(t, inodes.ExampleAliceRoot.LastModifiedAt(), stats.ModTime())
		assert.Equal(t, inodes.ExampleAliceFile.Size(), stats.Size())
		assert.Nil(t, stats.Sys())
		assert.Equal(t, stdfs.FileMode(0o660), stats.Mode().Perm())
		assert.False(t, stats.Mode().IsDir())
		assert.True(t, stats.Mode().IsRegular())
	})

	t.Run("OpenFile an existing file with write perms and without O_TRUNC", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_WRONLY, 0o644)
		assert.Nil(t, res)
		assert.EqualError(t, err, "open foo/bar: invalid argument")
	})

	t.Run("OpenFile with an invalid path", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		res, err := dav.OpenFile(ctx, "/foo/bar", os.O_RDONLY, 0o644)
		assert.Nil(t, res)
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
	})

	t.Run("OpenFile with a file not found", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(nil, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_RDONLY, 0o644)
		assert.EqualError(t, err, "open foo/bar: file does not exist")
		assert.Nil(t, res)
	})

	t.Run("OpenFile with O_CREATE and a file not found", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(nil, nil).Once()
		folderFSMock.On("CreateFile", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_WRONLY|os.O_CREATE, 0o644)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("OpenFile with O_EXCL with an existing file", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		assert.EqualError(t, err, "open foo/bar: file already exists")
		assert.ErrorIs(t, err, stdfs.ErrExist)
		assert.Nil(t, res)
	})

	t.Run("OpenFile with O_SYNC fails", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_WRONLY|os.O_SYNC, 0o644)
		assert.EqualError(t, err, "invalid argument: O_SYNC and O_APPEND not supported")
		assert.ErrorIs(t, err, stdfs.ErrInvalid)
		assert.Nil(t, res)
	})

	t.Run("OpenFile with O_APPEND fails", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_WRONLY|os.O_APPEND, 0o644)
		assert.EqualError(t, err, "invalid argument: O_SYNC and O_APPEND not supported")
		assert.ErrorIs(t, err, stdfs.ErrInvalid)
		assert.Nil(t, res)
	})

	t.Run("OpenFile with O_TRUNC success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		fsServiceMock := fs.NewMockService(t)
		dav := davFS{foldersMock, fsServiceMock}
		ctx := context.WithValue(context.Background(), sessionKeyCtx, &davsessions.ExampleAliceSession)

		foldersMock.On("GetUserFolder", mock.Anything, davsessions.ExampleAliceSession.UserID(), davsessions.ExampleAliceSession.FoldersIDs()[0]).
			Return(&folders.ExampleAlicePersonalFolder, nil).Once()

		folderFSMock := fs.NewMockFS(t)
		fsServiceMock.On("GetFolderFS", &folders.ExampleAlicePersonalFolder).Return(folderFSMock).Once()

		folderFSMock.On("Get", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()
		folderFSMock.On("RemoveAll", mock.Anything, "foo/bar").Return(nil).Once()
		folderFSMock.On("CreateFile", mock.Anything, "foo/bar").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := dav.OpenFile(ctx, "foo/bar", os.O_WRONLY|os.O_TRUNC, 0o644)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
