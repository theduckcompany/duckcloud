package httprouter

import (
	"fmt"
	"net/http"

	"github.com/Peltoche/neurone/src/tools/logger"
)

// NewServeMux builds a ServeMux that will route requests
// to the given EchoHandler.
func NewServeMux(handlers []MuxHandler, log *logger.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	for _, handler := range handlers {
		handler.Register(mux)
		log.Info(fmt.Sprintf("Register %q", handler.String()))
	}

	return mux
}
