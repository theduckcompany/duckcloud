package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

type Logger struct {
	*slog.Logger
}

func NewSLogger() *Logger {
	return &Logger{slog.New(slog.NewTextHandler(os.Stderr, nil))}
}
