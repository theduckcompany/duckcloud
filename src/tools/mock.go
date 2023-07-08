package tools

import (
	"testing"

	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/neilotoole/slogt"
	"golang.org/x/exp/slog"
)

type Mock struct {
	ClockMock *clock.MockClock
	UUIDMock  *uuid.MockService
	LogTest   *slog.Logger
	resWriter response.Writer
	JWTMock   *jwt.MockParser
}

func NewMock(t *testing.T) *Mock {
	t.Helper()

	return &Mock{
		ClockMock: clock.NewMockClock(t),
		UUIDMock:  uuid.NewMockService(t),
		LogTest:   slogt.New(t),
		resWriter: response.Init(response.Config{
			PrettyRender: true,
			HotReload:    false,
		}, logger.NewNoop()),
		JWTMock: jwt.NewMockParser(t),
	}
}

// Clock implements App.
func (m *Mock) Clock() clock.Clock {
	return m.ClockMock
}

func (m *Mock) JWT() jwt.Parser {
	return m.JWTMock
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
