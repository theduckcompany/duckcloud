package httprouter

import (
	"fmt"
	"net/http"

	"golang.org/x/exp/slog"
)

// NewServeMux builds a ServeMux that will route requests
// to the given EchoHandler.
func NewServeMux(handlers []MuxHandler, log *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	for _, handler := range handlers {
		handler.Register(mux)
		log.Info(fmt.Sprintf("Register %q", handler.String()))
	}

	return mux
}
