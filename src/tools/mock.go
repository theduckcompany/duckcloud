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

type Mock struct {
	ClockMock    *clock.MockClock
	UUIDMock     *uuid.MockService
	LogTest      *slog.Logger
	PasswordMock *password.MockPassword
	resWriter    response.Writer
}

func NewMock(t *testing.T) *Mock {
	t.Helper()

	return &Mock{
		ClockMock:    clock.NewMockClock(t),
		UUIDMock:     uuid.NewMockService(t),
		LogTest:      slogt.New(t),
		PasswordMock: password.NewMockPassword(t),
		resWriter: response.Init(response.Config{
			PrettyRender: true,
			HotReload:    false,
		}, logger.NewNoop()),
	}
}

// Clock implements App.
func (m *Mock) Clock() clock.Clock {
	return m.ClockMock
}

// UUID implements App.
func (m *Mock) UUID() uuid.Service {
	return m.UUIDMock
}

// Logger implements App.
func (m *Mock) Logger() *slog.Logger {
	return m.LogTest
}

func (m *Mock) ResWriter() response.Writer {
	return m.resWriter
}

func (m *Mock) Password() password.Password {
	return m.PasswordMock
}
