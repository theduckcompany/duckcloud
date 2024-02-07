package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/settings/general"
)

type renderUsersCmd struct {
	User    *users.User
	Session *websessions.Session
	Error   error
}

type Handler struct {
	html        html.Writer
	webSessions websessions.Service
	davSessions davsessions.Service
	spaces      spaces.Service
	users       users.Service
	uuid        uuid.Service
	auth        *auth.Authenticator
	tools       tools.Tools
}

func NewHandler(
	tools tools.Tools,
	html html.Writer,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	spaces spaces.Service,
	users users.Service,
	authent *auth.Authenticator,
) *Handler {
	return &Handler{
		html:        html,
		webSessions: webSessions,
		davSessions: davSessions,
		spaces:      spaces,
		users:       users,
		uuid:        tools.UUID(),
		auth:        authent,
		tools:       tools,
	}
}

func (h *Handler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/settings", http.RedirectHandler("/settings/general", http.StatusMovedPermanently).ServeHTTP)
	r.Get("/settings/general", h.getGeneralPage)

	newSecurityPage(h.tools, h.html, h.webSessions, h.davSessions, h.spaces, h.users, h.auth).Register(r, mids)
	newUsersPage(h.tools, h.html, h.users, h.auth).Register(r, mids)
}

type passwordFormCmd struct {
	Error error
}

func (h *Handler) getGeneralPage(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &general.LayoutTemplate{
		IsAdmin: user.IsAdmin(),
	})
}
