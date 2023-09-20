package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/web/html"
)

type homeHandler struct {
	html html.Writer
	auth *Authenticator
}

func newHomeHandler(
	html html.Writer,
	auth *Authenticator,
) *homeHandler {
	return &homeHandler{html, auth}
}

func (h *homeHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/", h.getHome)
}

func (h *homeHandler) String() string {
	return "web.home"
}

func (h *homeHandler) getHome(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "home/home.tmpl", map[string]interface{}{})
}
