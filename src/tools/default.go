package tools

import (
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/exp/slog"
)

type Default struct {
	clock     clock.Clock
	uuid      uuid.Service
	log       *slog.Logger
	resWriter response.Writer
	jwt       jwt.Parser
}

func (d *Default) JWT() jwt.Parser {
	return d.jwt
}

// Clock implements App.
//
// Return a clock.Default.
func (d *Default) Clock() clock.Clock {
	return d.clock
}

// UUID implements App.
//
// Return a *uuid.Default.
func (d *Default) UUID() uuid.Service {
	return d.uuid
}

// Logger implements App.
//
// Return a *logging.StdLogger.
func (d *Default) Logger() *slog.Logger {
	return d.log
}

// ResWriter implements App.
//
// Return a *response.Writer.
func (d *Default) ResWriter() response.Writer {
	return d.resWriter
}
