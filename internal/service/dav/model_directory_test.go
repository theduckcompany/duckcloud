package dav

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func Test_Directory(t *testing.T) {
	t.Run("Write is not implemented", func(t *testing.T) {
		duckDir := NewDirectory(nil, nil, "/foo")
		res, err := duckDir.Write([]byte{})
		assert.ErrorIs(t, err, fs.ErrInvalid)
		assert.Empty(t, res)
	})

	t.Run("Read is not implemented", func(t *testing.T) {
		duckDir := NewDirectory(nil, nil, "/foo")
		res, err := duckDir.Read([]byte{})
		assert.ErrorIs(t, err, fs.ErrInvalid)
		assert.Empty(t, res)
	})

	t.Run("Seek is not implemented", func(t *testing.T) {
		duckDir := NewDirectory(nil, nil, "/foo")
		res, err := duckDir.Seek(20, 20)
		assert.ErrorIs(t, err, fs.ErrInvalid)
		assert.Empty(t, res)
	})

	t.Run("Close does nothing", func(t *testing.T) {
		duckDir := NewDirectory(nil, nil, "/foo")
		err := duckDir.Close()
		assert.NoError(t, err)
	})

	t.Run("Stat returns the inode", func(t *testing.T) {
		duckDir := NewDirectory(&dfs.ExampleAliceRoot, nil, "/foo")
		res, err := duckDir.Stat()
		assert.NoError(t, err)
		assert.Equal(t, &dfs.ExampleAliceRoot, res)
	})

	t.Run("Readdir", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckDir := NewDirectory(&dfs.ExampleAliceRoot, fsMock, "/foo")

		fsMock.On("ListDir", mock.Anything, "/foo", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      2,
		}).Return([]dfs.INode{dfs.ExampleAliceFile, dfs.ExampleAliceFile}, nil).Once()

		res, err := duckDir.Readdir(2)
		assert.NoError(t, err)
		assert.Equal(t, []fs.FileInfo{&dfs.ExampleAliceFile, &dfs.ExampleAliceFile}, res)
	})

	t.Run("Readdir success with a response than the limit", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckDir := NewDirectory(&dfs.ExampleAliceRoot, fsMock, "/foo")

		fsMock.On("ListDir", mock.Anything, "/foo", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      2,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		// We expect 2 result but there is only 1 so io.EOF is return
		res, err := duckDir.Readdir(2)
		assert.ErrorIs(t, err, io.EOF)
		assert.Equal(t, []fs.FileInfo{&dfs.ExampleAliceFile}, res)
	})

	t.Run("Readdir several time success", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckDir := NewDirectory(&dfs.ExampleAliceRoot, fsMock, "/foo")

		fsMock.On("ListDir", mock.Anything, "/foo", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      1,
		}).Return([]dfs.INode{dfs.ExampleAliceFile}, nil).Once()

		res, err := duckDir.Readdir(1)
		assert.NoError(t, err)
		assert.Equal(t, []fs.FileInfo{&dfs.ExampleAliceFile}, res)

		// Call the second time. This will change the pagination
		fsMock.On("ListDir", mock.Anything, "/foo", &storage.PaginateCmd{
			StartAfter: map[string]string{"name": dfs.ExampleAliceFile.Name()},
			Limit:      1,
		}).Return([]dfs.INode{}, nil).Once()

		res, err = duckDir.Readdir(1)
		assert.ErrorIs(t, err, io.EOF)
		assert.Equal(t, []fs.FileInfo{}, res)
	})
}
