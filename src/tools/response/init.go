package response

import (
	"fmt"
	"net/http"
	"os"
	"path"

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
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to fetch the current workind dir: %s", err))
	}

	opts := render.Options{
		Directory:     path.Join(dir, "assets/html"),
		Layout:        "layout.html",
		IsDevelopment: cfg.HotReload,
		Extensions:    []string{".tmpl", ".html"},
	}

	if cfg.PrettyRender {
		opts.IndentJSON = true
		opts.IndentXML = true
	}

	return New(log, render.New(opts))
}
