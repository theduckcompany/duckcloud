package web

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
)

type homeHandler struct {
	response    response.Writer
	webSessions websessions.Service
	users       users.Service
}

func newHomeHandler(
	tools tools.Tools,
	webSessions websessions.Service,
	users users.Service,
) *homeHandler {
	return &homeHandler{
		response:    tools.ResWriter(),
		webSessions: webSessions,
		users:       users,
	}
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
	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	h.response.WriteHTML(w, r, http.StatusOK, "home/home.tmpl", map[string]interface{}{})
}

func (h *homeHandler) getUserAndSession(w http.ResponseWriter, r *http.Request) (*users.User, *websessions.Session) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return nil, nil
	}

	user, err := h.users.GetByID(ctx, currentSession.UserID())
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return nil, nil
	}

	if user == nil {
		_ = h.webSessions.Logout(r, w)
		return nil, nil
	}

	return user, currentSession
}
