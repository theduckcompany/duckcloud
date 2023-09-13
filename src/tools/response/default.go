package response

import (
	"errors"
	"log/slog"
	"net/http"
	"path"

	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/unrolled/render"
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

	t.log.Error("request failed", slog.String("error", err.Error()))

	if !errors.As(err, &ierr) {
		ierr = errs.Unhandled(err).(*errs.Error)
	}

	if rerr := t.render.JSON(w, ierr.Code(), ierr); rerr != nil {
		t.log.Error("failed to render a json response error", slog.String("error", err.Error()))
	}
}

func (t *Default) WriteHTML(w http.ResponseWriter, r *http.Request, status int, template string, args any) {
	layout := ""

	if r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == "" {
		layout = path.Join(path.Dir(template), "layout.tmpl")
	}

	if err := t.render.HTML(w, status, template, args, render.HTMLOptions{Layout: layout}); err != nil {
		t.log.Error("failed to render a json response", slog.String("error", err.Error()))
	}
}
