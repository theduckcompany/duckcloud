package response

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/unrolled/render"
)

type Config struct {
	PrettyRender bool `mapstructure:"prettyRender"`
	HotReload    bool `mapstructure:"hotReload"`
}

//go:generate mockery --name Writer
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
		Layout:        "layout.tmpl",
		IsDevelopment: cfg.HotReload,
		Extensions:    []string{".tmpl", ".html"},
	}

	if cfg.PrettyRender {
		opts.IndentJSON = true
		opts.IndentXML = true
	}

	return New(log, render.New(opts))
}
