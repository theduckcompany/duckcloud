package browser

import "github.com/theduckcompany/duckcloud/internal/tools/uuid"

type CreateDirTemplate struct {
	DirPath string
	SpaceID uuid.UUID
	Error   *string
}

func (t *CreateDirTemplate) Template() string { return "browser/modal_create_dir.tmpl" }
