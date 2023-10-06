package uploads

import (
	context "context"
	"errors"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrInvalidInodeID  = errors.New("invalid inode id")
	ErrInvalidFolderID = errors.New("invalid folder id")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, upload *Upload) error
	GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]Upload, error)
	Delete(ctx context.Context, uploadID uuid.UUID) error
}

type UploadService struct {
	storage Storage
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(storage Storage, tools tools.Tools) *UploadService {
	return &UploadService{storage, tools.Clock(), tools.UUID()}
}

func (s *UploadService) Register(ctx context.Context, cmd *RegisterUploadCmd) error {
	return s.storage.Save(ctx, &Upload{
		id:         s.uuid.New(),
		folderID:   cmd.FolderID,
		dir:        cmd.DirID,
		fileName:   cmd.FileName,
		fileID:     cmd.FileID,
		uploadedAt: s.clock.Now(),
	})
}

func (s *UploadService) Delete(ctx context.Context, upload *Upload) error {
	return s.storage.Delete(ctx, upload.id)
}

func (s *UploadService) GetOldest(ctx context.Context) (*Upload, error) {
	res, err := s.storage.GetAll(ctx, &storage.PaginateCmd{
		StartAfter: map[string]string{"uploaded_at": ""},
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil
	}

	return &res[0], nil
}
