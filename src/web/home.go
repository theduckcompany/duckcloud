package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
)

type homeHandler struct {
	response response.Writer
	auth     *Authenticator
}

func newHomeHandler(
	tools tools.Tools,
	auth *Authenticator,
) *homeHandler {
	return &homeHandler{response: tools.ResWriter(), auth: auth}
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

	h.response.WriteHTML(w, r, http.StatusOK, "home/home.tmpl", map[string]interface{}{})
}
