package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/home"
)

type HomePage struct {
	html html.Writer
	auth *auth.Authenticator
}

func NewHomePage(
	html html.Writer,
	auth *auth.Authenticator,
) *HomePage {
	return &HomePage{html, auth}
}

func (h *HomePage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.Defaults()...)
	}

	r.Get("/", h.getHome)
	r.Get("/logout", h.logout)
}

func (h *HomePage) logout(w http.ResponseWriter, r *http.Request) {
	h.auth.Logout(w, r)
}

func (h *HomePage) getHome(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &home.HomePageTmpl{})
}
