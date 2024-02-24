package logger

import (
	"io"
	"log/slog"
)

type Config struct {
	Output io.Writer
	Level  slog.Level `mapstructure:"level"`
}

func NewSLogger(cfg Config) *slog.Logger {
	return slog.New(slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
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
