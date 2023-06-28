package tools

import (
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/exp/slog"
)

type Tools struct {
	Clock     clock.Clock
	UUID      uuid.Service
	Log       *slog.Logger
	ResWriter response.Writer
	JWT       jwt.Parser
}

func Init(jwtCfg jwt.Config, log *slog.Logger) Tools {
	return Tools{
		Clock:     clock.New(),
		UUID:      uuid.NewProvider(),
		Log:       log,
		ResWriter: response.New(log),
		JWT:       jwt.NewDefault(jwtCfg),
	}
}
