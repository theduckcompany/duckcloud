package fileupload

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

func TestFileUploadJob(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		uploadsMock := uploads.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		tools := tools.NewMock(t)

		content := []byte("Hello, World!")

		afs := afero.NewMemMapFs()
		err := afero.WriteFile(afs, "some-file", content, 0o700)
		require.NoError(t, err)
		file, err := afs.Open("some-file")
		require.NoError(t, err)

		job := NewJob(foldersMock, uploadsMock, filesMock, inodesMock, tools)

		uploadsMock.On("GetOldest", mock.Anything).Return(&uploads.ExampleAliceUpload, nil).Once()

		filesMock.On("Open", mock.Anything, uploads.ExampleAliceUpload.FileID()).Return(file, nil).Once()
		inodesMock.On("CreateFile", mock.Anything, &inodes.CreateFileCmd{
			Parent:     uploads.ExampleAliceUpload.Dir(),
			Name:       uploads.ExampleAliceUpload.FileName(),
			Size:       uint64(len(content)),
			Checksum:   "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=", // SHA256 of "Hello, World!"
			FileID:     uploads.ExampleAliceUpload.FileID(),
			UploadedAt: uploads.ExampleAliceUpload.UploadedAt(),
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		// Start to update the size for all the parent folders
		inodesMock.On("GetByID", mock.Anything, *inodes.ExampleAliceFile.Parent()).Return(&inodes.ExampleAliceRoot, nil).Once()

		// Everything is correctly updated so we remove the task
		uploadsMock.On("Delete", mock.Anything, &uploads.ExampleAliceUpload).Return(nil).Once()

		// We check if there is some other upload.
		uploadsMock.On("GetOldest", mock.Anything).Return(nil, nil).Once()

		job.Run(ctx)
	})
}
