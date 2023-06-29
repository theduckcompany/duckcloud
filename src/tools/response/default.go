package response

import (
	"errors"
	"net/http"

	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/unrolled/render"
	"golang.org/x/exp/slog"
)

// Default is used to write the response into an http.ResponseWriter and log the error.
type Default struct {
	log    *slog.Logger
	render *render.Render
}

// New return a new Default.
func New(log *slog.Logger, render *render.Render) *Default {
	return &Default{log, render}
}

// Write the given res as a json body and statusCode.
func (t *Default) WriteJSON(w http.ResponseWriter, statusCode int, res any) {
	if err, ok := res.(error); ok {
		t.WriteJSONError(w, err)
		return
	}

	if err := t.render.JSON(w, statusCode, res); err != nil {
		t.log.Error("failed to render a json response", slog.String("error", err.Error()))
	}
}

// WriteJSONError write the given error into the ResponseWriter.
func (t *Default) WriteJSONError(w http.ResponseWriter, err error) {
	var ierr *errs.Error

	if !errors.As(err, &ierr) {
		ierr = errs.Unhandled(err).(*errs.Error)
	}

	if rerr := t.render.JSON(w, ierr.Code(), ierr); rerr != nil {
		t.log.Error("failed to render a json response error", slog.String("error", err.Error()))
	}
}

func (t *Default) WriteHTML(w http.ResponseWriter, status int, template string, args any) {
	if err := t.render.HTML(w, status, template, args); err != nil {
		t.log.Error("failed to render a json response", slog.String("error", err.Error()))
	}
}
