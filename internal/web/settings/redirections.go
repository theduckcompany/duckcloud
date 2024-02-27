package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
)

type Redirections struct{}

func NewRedirections() *Redirections {
	return &Redirections{}
}

func (h *Redirections) Register(r chi.Router, mids *router.Middlewares) {
	r.Get("/settings", http.RedirectHandler("/settings/security", http.StatusMovedPermanently).ServeHTTP)
}

type passwordFormCmd struct {
	Error error
}
