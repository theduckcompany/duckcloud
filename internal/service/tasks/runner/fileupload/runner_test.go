package fileupload

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	now                = time.Now().UTC()
	ExampleAliceUpload = scheduler.FileUploadArgs{
		FolderID:   folders.ExampleAlicePersonalFolder.ID(),
		Directory:  inodes.ExampleAliceRoot.ID(),
		FileName:   "foo.txt",
		FileID:     *inodes.ExampleAliceFile.FileID(),
		UploadedAt: now,
	}
)

func TestFileUploadTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewTaskRunner(nil, nil, nil)
		assert.Equal(t, model.FileUpload, job.Name())
	})

	t.Run("Run Success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		job := NewTaskRunner(foldersMock, filesMock, inodesMock)

		content := []byte("Hello, World!")
		afs := afero.NewMemMapFs()
		err := afero.WriteFile(afs, "some-file", content, 0o700)
		require.NoError(t, err)
		file, err := afs.Open("some-file")
		require.NoError(t, err)

		parsedTime, err := time.Parse(time.RFC3339, "2019-10-12T07:20:50.52Z")
		require.NoError(t, err)

		filesMock.On("Open", mock.Anything, uuid.UUID("01d39aea-9565-4e2f-9177-c3a2b4ea7ae9")).Return(file, nil).Once()
		inodesMock.On("CreateFile", mock.Anything, &inodes.CreateFileCmd{
			Parent:     uuid.UUID("c85dc5dc-daff-47da-a4e3-67690aae2de3"),
			Name:       "todo.txt",
			Size:       uint64(len(content)),
			Checksum:   "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=", // SHA256 of "Hello, World!"
			FileID:     uuid.UUID("01d39aea-9565-4e2f-9177-c3a2b4ea7ae9"),
			UploadedAt: parsedTime,
		}).Return(&inodes.ExampleAliceFile, nil).Once()

		// Start to update the size for all the parent folders
		inodesMock.On("RegisterWrite", mock.Anything, &inodes.ExampleAliceFile, int64(len(content)), inodes.ExampleAliceFile.LastModifiedAt()).
			Return(nil).Once()

		err = job.Run(ctx, json.RawMessage(`{
			"folder-id": "959a8808-273e-4079-90ed-a8394b356379",
			"directory": "c85dc5dc-daff-47da-a4e3-67690aae2de3",
			"file-name": "todo.txt",
			"file-id": "01d39aea-9565-4e2f-9177-c3a2b4ea7ae9",
			"uploaded-at": "2019-10-12T07:20:50.52Z"
		}`))
		assert.NoError(t, err)
	})

	t.Run("Run with some invalid json in args", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		job := NewTaskRunner(foldersMock, filesMock, inodesMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("RunArgs Success", func(t *testing.T) {
		foldersMock := folders.NewMockService(t)
		filesMock := files.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		job := NewTaskRunner(foldersMock, filesMock, inodesMock)

		content := []byte("Hello, World!")

		afs := afero.NewMemMapFs()
		err := afero.WriteFile(afs, "some-file", content, 0o700)
		require.NoError(t, err)
		file, err := afs.Open("some-file")
		require.NoError(t, err)

		filesMock.On("Open", mock.Anything, ExampleAliceUpload.FileID).Return(file, nil).Once()
		inodesMock.On("CreateFile", mock.Anything, &inodes.CreateFileCmd{
			Parent:     ExampleAliceUpload.Directory,
			Name:       ExampleAliceUpload.FileName,
			Size:       uint64(len(content)),
			Checksum:   "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=", // SHA256 of "Hello, World!"
			FileID:     ExampleAliceUpload.FileID,
			UploadedAt: ExampleAliceUpload.UploadedAt,
		}).Return(&inodes.ExampleAliceFile, nil).Once()
		inodesMock.On("RegisterWrite", mock.Anything, &inodes.ExampleAliceFile, int64(len(content)), inodes.ExampleAliceFile.LastModifiedAt()).
			Return(nil).Once()

		job.RunArgs(ctx, &ExampleAliceUpload)
	})
}
