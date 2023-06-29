package response

import (
	"net/http"

	"github.com/unrolled/render"
	"golang.org/x/exp/slog"
)

type Config struct {
	PrettyRender bool `mapstructure:"prettyRender"`
}

type Writer interface {
	WriteJSON(w http.ResponseWriter, statusCode int, res any)
	WriteJSONError(w http.ResponseWriter, err error)
	WriteHTML(w http.ResponseWriter, status int, template string, args any)
}

func Init(cfg Config, log *slog.Logger) Writer {
	opts := render.Options{
		Directory: "public/html",
		Layout:    "layout.html",
	}

	if cfg.PrettyRender {
		opts.IndentJSON = true
		opts.IndentXML = true
	}

	return New(log, render.New(opts))
}
