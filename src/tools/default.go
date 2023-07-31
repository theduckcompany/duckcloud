package tools

import (
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/password"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/exp/slog"
)

type Config struct {
	Response response.Config `mapstructure:"response"`
	Log      logger.Config   `mapstructure:"log"`
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
