package uploads

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Upload struct {
	id         uuid.UUID
	folderID   uuid.UUID
	dir        uuid.UUID
	fileName   string
	fileID     uuid.UUID
	uploadedAt time.Time
}

func (u *Upload) ID() uuid.UUID         { return u.id }
func (u *Upload) FolderID() uuid.UUID   { return u.folderID }
func (u *Upload) Dir() uuid.UUID        { return u.dir }
func (u *Upload) FileName() string      { return u.fileName }
func (u *Upload) FileID() uuid.UUID     { return u.fileID }
func (u *Upload) UploadedAt() time.Time { return u.uploadedAt }

type RegisterUploadCmd struct {
	FolderID uuid.UUID
	DirID    uuid.UUID
	FileName string
	FileID   uuid.UUID
}
