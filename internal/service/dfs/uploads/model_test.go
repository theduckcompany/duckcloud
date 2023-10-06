package uploads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadGetter(t *testing.T) {
	assert.Equal(t, ExampleAliceUpload.id, ExampleAliceUpload.ID())
	assert.Equal(t, ExampleAliceUpload.folderID, ExampleAliceUpload.FolderID())
	assert.Equal(t, ExampleAliceUpload.dir, ExampleAliceUpload.Dir())
	assert.Equal(t, ExampleAliceUpload.fileName, ExampleAliceUpload.FileName())
	assert.Equal(t, ExampleAliceUpload.fileID, ExampleAliceUpload.FileID())
	assert.Equal(t, ExampleAliceUpload.uploadedAt, ExampleAliceUpload.UploadedAt())
}
