package html

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/unrolled/render"
)

//go:embed *
var embeddedTemplates embed.FS

type Config struct {
	PrettyRender bool `mapstructure:"prettyRender"`
	HotReload    bool `mapstructure:"hotReload"`
}

//go:generate mockery --name Writer
type Writer interface {
	WriteHTML(w http.ResponseWriter, r *http.Request, status int, template string, args any)
	WriteHTMLErrorPage(w http.ResponseWriter, r *http.Request, err error)
}

type Renderer struct {
	render *render.Render
}

func NewRenderer(cfg Config) *Renderer {
	var directory string
	var fs render.FileSystem

	if cfg.HotReload {
		dir, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("failed to fetch the current workind dir: %s", err))
		}

		directory = path.Join(dir, "src/web/html/templates")
		fs = render.LocalFileSystem{}
	} else {
		directory = ""
		fs = &render.EmbedFileSystem{FS: embeddedTemplates}
	}

	opts := render.Options{
		Directory:     directory,
		FileSystem:    fs,
		Layout:        "",
		IsDevelopment: cfg.HotReload,
		Extensions:    []string{".tmpl", ".html"},
		Funcs: []template.FuncMap{
			{
				"humanTime": humanize.Time,
				"humanDate": func(t time.Time) string { return t.Format(time.DateTime) },
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
				"getInodeIconClass": func(_ string, isDir bool) string {
					if isDir {
						return "bi-folder-fill text-primary"
					}

					return "bi-file-earmark-fill text-muted"
				},
			},
		},
	}

	if cfg.PrettyRender {
		opts.IndentXML = true
	}

	renderer := render.New(opts)
	renderer.CompileTemplates()

	return &Renderer{renderer}
}

func (t *Renderer) WriteHTML(w http.ResponseWriter, r *http.Request, status int, template string, args any) {
	layout := ""

	if r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == "" {
		layout = path.Join(path.Dir(template), "layout.tmpl")
	}

	if err := t.render.HTML(w, status, template, args, render.HTMLOptions{Layout: layout}); err != nil {
		logger.LogEntrySetAttrs(r, slog.String("render-error", err.Error()))
	}
}

func (t *Renderer) WriteHTMLErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	layout := ""

	reqID := r.Context().Value(middleware.RequestIDKey).(string)

	if r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == "" {
		layout = path.Join("home/layout.tmpl")
	}

	logger.LogEntrySetError(r, err)

	if err := t.render.HTML(w, http.StatusInternalServerError, "home/500.tmpl", map[string]any{
		"requestID": reqID,
	}, render.HTMLOptions{Layout: layout}); err != nil {
		logger.LogEntrySetAttrs(r, slog.String("render-error", err.Error()))
	}
}
