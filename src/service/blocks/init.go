package blocks

import (
	"context"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/spf13/afero"
)

type Config struct {
	Path string `json:"path"`
}

type Service interface {
	Open(ctx context.Context, inodeID uuid.UUID) (afero.File, error)
}

func Init(cfg Config, tools tools.Tools) (Service, error) {
	fs := afero.NewOsFs()

	return NewFSService(fs, cfg.Path, tools.Logger())
}
