package logger

import (
	"os"

	"golang.org/x/exp/slog"
)

func NewSLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		// Remove default time slog.Attr. It will be replaced by the one
		// from the router middleware.
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
}
