package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type Handler struct {
	html        html.Writer
	webSessions websessions.Service
	davSessions davsessions.Service
	spaces      spaces.Service
	users       users.Service
	uuid        uuid.Service
	scheduler   scheduler.Service
	auth        *auth.Authenticator
	tools       tools.Tools
}

func NewHandler(
	tools tools.Tools,
	html html.Writer,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	spaces spaces.Service,
	scheduler scheduler.Service,
	users users.Service,
	authent *auth.Authenticator,
) *Handler {
	return &Handler{
		html:        html,
		webSessions: webSessions,
		davSessions: davSessions,
		spaces:      spaces,
		scheduler:   scheduler,
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

	r.Get("/settings", http.RedirectHandler("/settings/security", http.StatusMovedPermanently).ServeHTTP)

	newSecurityPage(h.tools, h.html, h.webSessions, h.davSessions, h.spaces, h.users, h.auth).Register(r, mids)
	newUsersPage(h.tools, h.html, h.users, h.auth).Register(r, mids)
	newSpacesPage(h.html, h.spaces, h.users, h.auth, h.scheduler, h.tools).Register(r, mids)
}

type passwordFormCmd struct {
	Error error
}
