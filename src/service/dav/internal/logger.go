package internal

import (
	"net/http"

	"golang.org/x/exp/slog"
)

func NewLogger(log *slog.Logger) func(r *http.Request, err error) {
	return func(r *http.Request, err error) {
		if err != nil {
			log.WithGroup("dav").
				ErrorCtx(r.Context(), "dav error", slog.String("error", err.Error()))
		}
	}
}
