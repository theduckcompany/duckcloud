package tools

import (
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/exp/slog"
)

// Tools regroup all the utilities required for a working server.
type Tools interface {
	Clock() clock.Clock
	UUID() uuid.Service
	Logger() *slog.Logger
	ResWriter() response.Writer
	JWT() jwt.Parser
}

func Init(jwtCfg jwt.Config, log *slog.Logger) Default {
	return Default{
		clock:     clock.NewDefault(),
		uuid:      uuid.NewProvider(),
		log:       log,
		resWriter: response.New(log),
		jwt:       jwt.NewDefault(jwtCfg),
	}
}
