package uploads

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_UploadService(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC()

	t.Run("Register", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-uuid"))
		tools.ClockMock.On("Now").Return(now)

		storageMock.On("Save", mock.Anything, &Upload{
			id:         uuid.UUID("some-uuid"),
			folderID:   ExampleAliceUpload.folderID,
			dir:        ExampleAliceUpload.dir,
			fileName:   ExampleAliceUpload.fileName,
			fileID:     ExampleAliceUpload.fileID,
			uploadedAt: now,
		}).Return(nil)

		err := service.Register(ctx, &RegisterUploadCmd{
			FolderID: ExampleAliceUpload.folderID,
			DirID:    ExampleAliceUpload.dir,
			FileName: ExampleAliceUpload.fileName,
			FileID:   ExampleAliceUpload.fileID,
		})
		assert.NoError(t, err)
	})

	t.Run("GetOldest", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAll", mock.Anything, &storage.PaginateCmd{StartAfter: map[string]string{"uploaded_at": ""}, Limit: 1}).
			Return([]Upload{ExampleAliceUpload}, nil).Once()

		res, err := service.GetOldest(ctx)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceUpload, res)
	})

	t.Run("GetOldest without any upload", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAll", mock.Anything, &storage.PaginateCmd{StartAfter: map[string]string{"uploaded_at": ""}, Limit: 1}).
			Return([]Upload{}, nil).Once()

		res, err := service.GetOldest(ctx)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}
