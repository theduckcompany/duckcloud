package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/exp/slog"
)

// Default is used to write the response into an http.ResponseWriter and log the error.
type Default struct {
	log *slog.Logger
}

// New return a new Default.
func New(log *slog.Logger) *Default {
	return &Default{log}
}

// Write the given res as a json body and statusCode.
func (t *Default) Write(w http.ResponseWriter, r *http.Request, res any, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if res != nil {
		_ = json.NewEncoder(w).Encode(res)
		t.log.WithGroup("http").ErrorCtx(
			r.Context(),
			"",
			slog.Int("status", statusCode),
		)
	}
}

// WriteError write the given error into the ResponseWriter.
func (t *Default) WriteError(err error, w http.ResponseWriter, r *http.Request) {
	var ierr *Error

	if errors.As(err, &ierr) {
		_ = json.NewEncoder(w).Encode(ierr)
		w.WriteHeader(ierr.Code)
		t.log.WithGroup("http").ErrorCtx(
			r.Context(),
			"",
			slog.Int("status", ierr.Code),
			slog.String("error", ierr.Internal.Error()),
		)
	}
}
