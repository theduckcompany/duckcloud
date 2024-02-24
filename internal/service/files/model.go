package files

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FileMeta struct {
	uploadedAt time.Time
	key        *secret.SealedKey
	id         uuid.UUID
	mimetype   string
	checksum   string
	size       uint64
}

func (f *FileMeta) ID() uuid.UUID         { return f.id }
func (f *FileMeta) Size() uint64          { return f.size }
func (f *FileMeta) MimeType() string      { return f.mimetype }
func (f *FileMeta) Checksum() string      { return f.checksum }
func (f *FileMeta) UploadedAt() time.Time { return f.uploadedAt }
