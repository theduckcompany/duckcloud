package fs

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

func Test_Directory(t *testing.T) {
	t.Run("Directory implements ReadDirFile", func(t *testing.T) {
		assert.Implements(t, (*fs.ReadDirFile)(nil), new(Directory))
	})

	t.Run("Directory implements but doesn't support some ReadDirFile methods", func(t *testing.T) {
		dirPath := t.TempDir()

		dir, err := os.Open(dirPath)
		require.NoError(t, err)

		duckDir := NewDirectory(nil, nil, &inodes.PathCmd{FullName: dirPath})

		t.Run("Write", func(t *testing.T) {
			res, err := duckDir.Write([]byte{})
			expectedRes, expectedErr := dir.Write([]byte{})

			assert.Equal(t, expectedErr, err, "err")
			assert.Equal(t, expectedRes, res, "res")
		})

		t.Run("Read", func(t *testing.T) {
			res, err := duckDir.Read([]byte{})
			expectedRes, expectedErr := dir.Read([]byte{})

			assert.Equal(t, expectedErr, err, "err")
			assert.Equal(t, expectedRes, res, "res")
		})
	})
}
