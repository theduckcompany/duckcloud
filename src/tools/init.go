package tools

import (
	"log/slog"

	"github.com/myminicloud/myminicloud/src/tools/clock"
	"github.com/myminicloud/myminicloud/src/tools/password"
	"github.com/myminicloud/myminicloud/src/tools/response"
	"github.com/myminicloud/myminicloud/src/tools/uuid"
)

// Tools regroup all the utilities required for a working server.
type Tools interface {
	Clock() clock.Clock
	UUID() uuid.Service
	Logger() *slog.Logger
	ResWriter() response.Writer
	Password() password.Password
}
