package logger

import (
	"io"

	"golang.org/x/exp/slog"
)

func NewNoop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
