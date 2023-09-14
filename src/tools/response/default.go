package response

import (
	"errors"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/logger"
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
		logger.LogEntrySetField(r, "render-error", err.Error())
	}
}

// WriteJSONError write the given error into the ResponseWriter.
func (t *Default) WriteJSONError(w http.ResponseWriter, r *http.Request, err error) {
	var ierr *errs.Error

	logger.LogEntrySetField(r, "error", err.Error())

	if !errors.As(err, &ierr) {
		ierr = errs.Unhandled(err).(*errs.Error)
	}

	if rerr := t.render.JSON(w, ierr.Code(), ierr); rerr != nil {
		logger.LogEntrySetField(r, "render-error", err.Error())
	}
}

func (t *Default) WriteHTML(w http.ResponseWriter, r *http.Request, status int, template string, args any) {
	layout := ""

	if r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == "" {
		layout = path.Join(path.Dir(template), "layout.tmpl")
	}

	if err := t.render.HTML(w, status, template, args, render.HTMLOptions{Layout: layout}); err != nil {
		logger.LogEntrySetField(r, "render-error", err.Error())
	}
}

func (t *Default) WriteHTMLErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	layout := ""

	reqID := r.Context().Value(middleware.RequestIDKey).(string)

	if r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == "" {
		layout = path.Join("home/layout.tmpl")
	}

	logger.LogEntrySetField(r, "error", err.Error())

	if err := t.render.HTML(w, http.StatusInternalServerError, "home/500.tmpl", map[string]any{
		"requestID": reqID,
	}, render.HTMLOptions{Layout: layout}); err != nil {
		logger.LogEntrySetField(r, "render-error", err.Error())
	}
}
