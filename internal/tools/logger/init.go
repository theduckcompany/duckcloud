package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Level slog.Level `mapstructure:"level"`
}

func NewSLogger(cfg Config) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.Level,
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
