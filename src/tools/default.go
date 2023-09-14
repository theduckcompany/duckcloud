package tools

import (
	"log/slog"
	"testing"

	"github.com/neilotoole/slogt"
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

type Toolbox struct {
	clock     clock.Clock
	uuid      uuid.Service
	resWriter response.Writer
	log       *slog.Logger
	password  password.Password
}

func NewToolbox(cfg Config) *Toolbox {
	log := logger.NewSLogger(cfg.Log)

	return &Toolbox{
		clock:     clock.NewDefault(),
		uuid:      uuid.NewProvider(),
		log:       log,
		resWriter: response.Init(cfg.Response),
		password:  password.NewBcryptPassword(),
	}
}

func NewToolboxForTest(t *testing.T) *Toolbox {
	t.Helper()

	log := slogt.New(t)
	return &Toolbox{
		clock:     clock.NewDefault(),
		uuid:      uuid.NewProvider(),
		log:       log,
		resWriter: response.Init(response.Config{PrettyRender: true, HotReload: false}),
		password:  password.NewBcryptPassword(),
	}
}

// Clock implements App.
//
// Return a clock.Default.
func (d *Toolbox) Clock() clock.Clock {
	return d.clock
}

// UUID implements App.
//
// Return a *uuid.Default.
func (d *Toolbox) UUID() uuid.Service {
	return d.uuid
}

// Logger implements App.
//
// Return a *logging.StdLogger.
func (d *Toolbox) Logger() *slog.Logger {
	return d.log
}

// ResWriter implements App.
//
// Return a *response.Writer.
func (d *Toolbox) ResWriter() response.Writer {
	return d.resWriter
}

func (d *Toolbox) Password() password.Password {
	return d.password
}
