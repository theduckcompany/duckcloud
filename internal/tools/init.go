package tools

import (
	"log/slog"

	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/password"
	"github.com/theduckcompany/duckcloud/internal/tools/response"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

// Tools regroup all the utilities required for a working server.
type Tools interface {
	Clock() clock.Clock
	UUID() uuid.Service
	Logger() *slog.Logger
	ResWriter() response.Writer
	Password() password.Password
}
