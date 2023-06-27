package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

func NewSLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}
