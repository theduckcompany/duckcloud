package internal

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools/logger"
	"golang.org/x/exp/slog"
)

func NewLogger(log *logger.Logger) func(r *http.Request, err error) {
	return func(r *http.Request, err error) {
		if err != nil {
			log.WithGroup("dav").
				ErrorCtx(r.Context(), "dav error", slog.String("error", err.Error()))
		}
	}
}
