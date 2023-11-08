package files

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var ExampleAliceUpload = scheduler.FileUploadArgs{
	FolderID:   folders.ExampleAlicePersonalFolder.ID(),
	INodeID:    uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	FileID:     ExampleFile1.ID(),
	UploadedAt: now,
}

type file struct {
	*bytes.Reader
}

func (f file) Close() error { return nil }

func TestFileUploadTask(t *testing.T) {
	ctx := context.Background()

	t.Run("Name", func(t *testing.T) {
		job := NewFileUploadTaskRunner(nil, nil, nil)
		assert.Equal(t, "file-upload", job.Name())
	})

	t.Run("Run Success", func(t *testing.T) {
		filesMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFileUploadTaskRunner(filesMock, storageMock, schedulerMock)

		// text/plain content type
		content := []byte("Hello, World!")

		file := file{bytes.NewReader(content)}

		parsedTime, err := time.Parse(time.RFC3339, "2019-10-12T07:20:50.52Z")
		require.NoError(t, err)

		filesMock.On("Download", mock.Anything, uuid.UUID("01d39aea-9565-4e2f-9177-c3a2b4ea7ae9")).Return(file, nil).Once()

		storageMock.On("GetByChecksum", mock.Anything, "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=").Return(nil, errNotFound).Once()
		storageMock.On("Save", mock.Anything, &FileMeta{
			id:         uuid.UUID("01d39aea-9565-4e2f-9177-c3a2b4ea7ae9"),
			size:       uint64(len(content)),
			mimetype:   "text/plain; charset=utf-8",
			checksum:   "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=",
			uploadedAt: parsedTime,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			ModifiedAt: parsedTime,
		}).Return(nil).Once()

		err = job.Run(ctx, json.RawMessage(`{
			"folder-id": "959a8808-273e-4079-90ed-a8394b356379",
			"file-id": "01d39aea-9565-4e2f-9177-c3a2b4ea7ae9",
			"inode-id": "f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f",
			"uploaded-at": "2019-10-12T07:20:50.52Z"
		}`))
		assert.NoError(t, err)
	})

	t.Run("Run with some invalid json in args", func(t *testing.T) {
		filesMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFileUploadTaskRunner(filesMock, storageMock, schedulerMock)

		err := job.Run(ctx, json.RawMessage(`some-invalid-json`))
		assert.ErrorContains(t, err, "failed to unmarshal the args")
	})

	t.Run("RunArgs Success", func(t *testing.T) {
		filesMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFileUploadTaskRunner(filesMock, storageMock, schedulerMock)

		content := []byte("Hello, World!")

		file := file{bytes.NewReader(content)}
		filesMock.On("Download", mock.Anything, ExampleAliceUpload.FileID).Return(file, nil).Once()

		storageMock.On("GetByChecksum", mock.Anything, "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=").Return(nil, errNotFound).Once()
		storageMock.On("Save", mock.Anything, &FileMeta{
			id:         ExampleAliceUpload.FileID,
			size:       uint64(len(content)),
			mimetype:   "text/plain; charset=utf-8",
			checksum:   "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=",
			uploadedAt: ExampleFile1.uploadedAt,
		}).Return(nil).Once()
		schedulerMock.On("RegisterFSRefreshSizeTask", mock.Anything, &scheduler.FSRefreshSizeArg{
			INode:      ExampleAliceUpload.INodeID,
			ModifiedAt: ExampleFile1.uploadedAt,
		}).Return(nil).Once()

		job.RunArgs(ctx, &ExampleAliceUpload)
	})

	t.Run("RunArgs with the same file already uploaded", func(t *testing.T) {
		filesMock := NewMockService(t)
		storageMock := NewMockStorage(t)
		schedulerMock := scheduler.NewMockService(t)
		job := NewFileUploadTaskRunner(filesMock, storageMock, schedulerMock)

		content := []byte("Hello, World!")

		file := file{bytes.NewReader(content)}
		filesMock.On("Download", mock.Anything, ExampleAliceUpload.FileID).Return(file, nil).Once()

		storageMock.On("GetByChecksum", mock.Anything, "3_1gIbsr1bCvZ2KQgJ7DpTGR3YHH9wpLKGiKNiGCmG8=").Return(&ExampleFile1, nil).Once()
		schedulerMock.On("RegisterFSRemoveDuplicateFile", mock.Anything, &scheduler.FSRemoveDuplicateFileArgs{
			INode:        ExampleAliceUpload.INodeID,
			TargetFileID: ExampleAliceUpload.FileID,
		}).Return(nil).Once()

		job.RunArgs(ctx, &ExampleAliceUpload)
	})
}
