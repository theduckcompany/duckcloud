package tools

import (
	"log/slog"

	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/logger"
	"github.com/theduckcompany/duckcloud/src/tools/password"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type Config struct {
	Response response.Config `json:"response"`
	Log      logger.Config   `json:"log"`
}

type Prod struct {
	clock     clock.Clock
	uuid      uuid.Service
	resWriter response.Writer
	log       *slog.Logger
	password  password.Password
}

func NewToolbox(cfg Config) *Prod {
	log := logger.NewSLogger(cfg.Log)

	return &Prod{
		clock:     clock.NewDefault(),
		uuid:      uuid.NewProvider(),
		log:       log,
		resWriter: response.Init(cfg.Response, log),
		password:  password.NewBcryptPassword(),
	}
}

// Clock implements App.
//
// Return a clock.Default.
func (d *Prod) Clock() clock.Clock {
	return d.clock
}

// UUID implements App.
//
// Return a *uuid.Default.
func (d *Prod) UUID() uuid.Service {
	return d.uuid
}

// Logger implements App.
//
// Return a *logging.StdLogger.
func (d *Prod) Logger() *slog.Logger {
	return d.log
}

// ResWriter implements App.
//
// Return a *response.Writer.
func (d *Prod) ResWriter() response.Writer {
	return d.resWriter
}

func (d *Prod) Password() password.Password {
	return d.password
}
