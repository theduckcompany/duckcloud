package response

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"text/template"

	"github.com/dustin/go-humanize"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
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
	WriteHTML(w http.ResponseWriter, r *http.Request, status int, template string, args any)
}

func Init(cfg Config, log *slog.Logger) Writer {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to fetch the current workind dir: %s", err))
	}

	opts := render.Options{
		Directory:     path.Join(dir, "assets/html"),
		Layout:        "",
		IsDevelopment: cfg.HotReload,
		Extensions:    []string{".tmpl", ".html"},
		Funcs: []template.FuncMap{
			{
				"humanTime": humanize.Time,
				"humanSize": humanize.Bytes,
			},
			{
				"pathJoin": func(elems ...any) string {
					strElems := make([]string, len(elems))
					for i, elem := range elems {
						switch elem := elem.(type) {
						case uuid.UUID:
							strElems[i] = string(elem)
						default:
							strElems[i] = elem.(string)
						}
					}
					return path.Join(strElems...)
				},
				"getInodeIconClass": func(name string, isDir bool) string {
					if isDir {
						return "bi-folder-fill text-primary"
					}

					return "bi-file-earmark-fill text-muted"
				},
			},
		},
	}

	if cfg.PrettyRender {
		opts.IndentJSON = true
		opts.IndentXML = true
	}

	return New(log, render.New(opts))
}
