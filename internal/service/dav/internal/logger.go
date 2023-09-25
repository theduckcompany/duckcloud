package internal

import (
	"log/slog"
	"net/http"
)

func NewLogger(log *slog.Logger) func(r *http.Request, err error) {
	return func(r *http.Request, err error) {
		if err != nil {
			log.WithGroup("dav").
				ErrorContext(r.Context(), "dav error", slog.String("error", err.Error()))
		}
	}
}
