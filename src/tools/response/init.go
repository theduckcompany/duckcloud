package response

import (
	"net/http"

	"github.com/unrolled/render"
	"golang.org/x/exp/slog"
)

type Config struct {
	PrettyRender bool `mapstructure:"prettyRender"`
	HotReload    bool `mapstructure:"hotReload"`
}

type Writer interface {
	WriteJSON(w http.ResponseWriter, statusCode int, res any)
	WriteJSONError(w http.ResponseWriter, err error)
	WriteHTML(w http.ResponseWriter, status int, template string, args any)
}

func Init(cfg Config, log *slog.Logger) Writer {
	opts := render.Options{
		Directory:     "assets/html",
		Layout:        "layout.html",
		IsDevelopment: cfg.HotReload,
	}

	if cfg.PrettyRender {
		opts.IndentJSON = true
		opts.IndentXML = true
	}

	return New(log, render.New(opts))
}
