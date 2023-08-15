package logger

import (
	"io"
	"log/slog"
)

func NewNoop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
