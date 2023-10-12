package dav

import (
	"bytes"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

type readSeekCloser struct {
	io.ReadSeeker
}

func (r *readSeekCloser) Close() error { return nil }

func Test_File(t *testing.T) {
	t.Run("Readdir is not implemented", func(t *testing.T) {
		duckFile := NewFile("", nil)
		res, err := duckFile.Readdir(2)
		assert.ErrorIs(t, err, fs.ErrInvalid)
		assert.Empty(t, res)
	})

	t.Run("ReadDir is not implemented", func(t *testing.T) {
		duckFile := NewFile("", nil)
		res, err := duckFile.ReadDir(2)
		assert.ErrorIs(t, err, fs.ErrInvalid)
		assert.Empty(t, res)
	})

	t.Run("Seek is not implemented", func(t *testing.T) {
		duckFile := NewFile("", nil)
		res, err := duckFile.Seek(2, 22)
		assert.ErrorIs(t, err, fs.ErrInvalid)
		assert.Empty(t, res)
	})

	t.Run("Stat success", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckFile := NewFile("/foo/bar.txt", fsMock)

		fsMock.On("Get", mock.Anything, "/foo/bar.txt").Return(&inodes.ExampleAliceFile, nil).Once()

		res, err := duckFile.Stat()
		assert.NoError(t, err)
		assert.Equal(t, &inodes.ExampleAliceFile, res)
	})

	t.Run("Write success", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckFile := NewFile("/foo/bar.txt", fsMock)

		content := []byte("Hello, World!")
		// waitEndTest := make(chan struct{})

		fsMock.On("Upload", mock.Anything, "/foo/bar.txt", mock.Anything).
			Run(func(args mock.Arguments) {
				uploaded, err := io.ReadAll(args[2].(io.Reader))
				require.NoError(t, err)
				require.Equal(t, content, uploaded)
				// waitEndTest <- struct{}{}
			}).Return(nil).Once()

		res, err := duckFile.Write(content)
		assert.NoError(t, err)
		assert.Equal(t, 13, res)

		err = duckFile.Close()
		assert.NoError(t, err)

		// _ = <-waitEndTest
	})

	t.Run("Several Write success", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckFile := NewFile("/foo/bar.txt", fsMock)

		fsMock.On("Upload", mock.Anything, "/foo/bar.txt", mock.Anything).
			Run(func(args mock.Arguments) {
				uploaded, err := io.ReadAll(args[2].(io.Reader))
				require.NoError(t, err)
				require.Equal(t, []byte("Hello, World!"), uploaded)
			}).Return(nil).Once()

		res, err := duckFile.Write([]byte("Hello, "))
		assert.NoError(t, err)
		assert.Equal(t, 7, res)

		res, err = duckFile.Write([]byte("World!"))
		assert.NoError(t, err)
		assert.Equal(t, 6, res)

		err = duckFile.Close()
		assert.NoError(t, err)
	})

	t.Run("Read success", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckFile := NewFile("/foo/bar.txt", fsMock)

		content := bytes.NewReader([]byte("Hello, World!"))

		fsMock.On("Download", mock.Anything, "/foo/bar.txt").
			Return(&readSeekCloser{content}, nil).Once()

		res, err := io.ReadAll(duckFile)
		assert.NoError(t, err)
		assert.Equal(t, []byte("Hello, World!"), res)

		err = duckFile.Close()
		assert.NoError(t, err)
	})

	t.Run("Several Read success", func(t *testing.T) {
		fsMock := dfs.NewMockFS(t)
		duckFile := NewFile("/foo/bar.txt", fsMock)

		content := bytes.NewReader([]byte("Hello, World!"))

		fsMock.On("Download", mock.Anything, "/foo/bar.txt").
			Return(&readSeekCloser{content}, nil).Once()

		body1 := make([]byte, 8)
		res, err := duckFile.Read(body1)
		assert.NoError(t, err)
		assert.Equal(t, 8, res)
		assert.Equal(t, []byte("Hello, W"), body1)

		body2 := make([]byte, 8)
		res, err = duckFile.Read(body2)
		assert.NoError(t, err)
		assert.Equal(t, 5, res)
		assert.EqualValues(t, []byte("orld!"), body2[0:5])

		err = duckFile.Close()
		assert.NoError(t, err)
	})
}
