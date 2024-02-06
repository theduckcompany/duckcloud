package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type homeHandler struct {
	html html.Writer
	auth *auth.Authenticator
}

func newHomeHandler(
	html html.Writer,
	auth *auth.Authenticator,
) *homeHandler {
	return &homeHandler{html, auth}
}

func (h *homeHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/", h.getHome)
	r.Get("/logout", h.logout)
}

func (h *homeHandler) String() string {
	return "web.home"
}

func (h *homeHandler) logout(w http.ResponseWriter, r *http.Request) {
	h.auth.Logout(w, r)
}

func (h *homeHandler) getHome(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "home/page", map[string]interface{}{})
}
