package tools

import (
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"go.uber.org/fx"
	"golang.org/x/exp/slog"
)

type Tools struct {
	fx.Out
	Clock     clock.Clock
	UUID      uuid.Service
	Log       *slog.Logger
	ResWriter response.Writer
	JWT       jwt.Parser
}

func Init(jwtCfg jwt.Config) Tools {
	log := logger.NewSLogger()

	return Tools{
		Clock:     clock.New(),
		UUID:      uuid.NewProvider(),
		Log:       log,
		ResWriter: response.New(log),
		JWT:       jwt.NewDefault(jwtCfg),
	}
}
