package response

import (
	"net/http"

	"github.com/unrolled/render"
)

type Config struct {
	PrettyRender bool `mapstructure:"prettyRender"`
}

//go:generate mockery --name Writer
type Writer interface {
	WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, res any)
	WriteJSONError(w http.ResponseWriter, r *http.Request, err error)
}

func Init(cfg Config) Writer {
	opts := render.Options{}

	if cfg.PrettyRender {
		opts.IndentJSON = true
		opts.IndentXML = true
	}

	return New(render.New(opts))
}
