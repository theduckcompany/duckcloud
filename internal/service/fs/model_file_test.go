package fs

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

func Test_File(t *testing.T) {
	t.Run("File implements ReadDirFile", func(t *testing.T) {
		assert.Implements(t, (*fs.ReadDirFile)(nil), new(File))
	})

	t.Run("File implements but doesn't support some ReadDirFile methods", func(t *testing.T) {
		tempFile := t.TempDir() + "/foo.txt"
		err := os.WriteFile(tempFile, []byte("Hello, World!"), 0o700)
		require.NoError(t, err)

		file, err := os.Open(tempFile)
		require.NoError(t, err)

		duckFile := &File{cmd: &inodes.PathCmd{FullName: tempFile}}

		t.Run("ReadDir", func(t *testing.T) {
			res, err := duckFile.ReadDir(0)
			expectedRes, expectedErr := file.ReadDir(0)

			assert.Equal(t, expectedErr, err)
			assert.Equal(t, expectedRes, res)
		})

		t.Run("Readdir", func(t *testing.T) {
			res2, err := duckFile.Readdir(0)
			expectedRes2, expectedErr := file.Readdir(0)

			assert.Equal(t, expectedErr, err)
			assert.Equal(t, expectedRes2, res2)
		})
	})
}
