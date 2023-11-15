package files

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FileMeta struct {
	id         uuid.UUID
	size       uint64
	mimetype   string
	checksum   string
	key        secret.Text
	uploadedAt time.Time
}

func (f *FileMeta) ID() uuid.UUID         { return f.id }
func (f *FileMeta) Size() uint64          { return f.size }
func (f *FileMeta) MimeType() string      { return f.mimetype }
func (f *FileMeta) Checksum() string      { return f.checksum }
func (f *FileMeta) UploadedAt() time.Time { return f.uploadedAt }
