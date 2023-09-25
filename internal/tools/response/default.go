package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/unrolled/render"
)

// Default is used to write the response into an http.ResponseWriter and log the error.
type Default struct {
	render *render.Render
}

// New return a new Default.
func New(render *render.Render) *Default {
	return &Default{render}
}

// Write the given res as a json body and statusCode.
func (t *Default) WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, res any) {
	if err, ok := res.(error); ok {
		t.WriteJSONError(w, r, err)
		return
	}

	if err := t.render.JSON(w, statusCode, res); err != nil {
		logger.LogEntrySetAttrs(r, slog.String("render-error", err.Error()))
	}
}

// WriteJSONError write the given error into the ResponseWriter.
func (t *Default) WriteJSONError(w http.ResponseWriter, r *http.Request, err error) {
	var ierr *errs.Error

	logger.LogEntrySetError(r, err)

	if !errors.As(err, &ierr) {
		ierr = errs.Unhandled(err).(*errs.Error)
	}

	if rerr := t.render.JSON(w, ierr.Code(), ierr); rerr != nil {
		logger.LogEntrySetAttrs(r, slog.String("render-error", err.Error()))
	}
}
