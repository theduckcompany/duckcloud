package web

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

type settingsHandler struct {
	response    response.Writer
	webSessions websessions.Service
	davSessions davsessions.Service
	users       users.Service
}

func newSettingsHandler(
	tools tools.Tools,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	users users.Service,
) *settingsHandler {
	return &settingsHandler{
		response:    tools.ResWriter(),
		webSessions: webSessions,
		davSessions: davSessions,
		users:       users,
	}
}

func (h *settingsHandler) Register(r chi.Router, mids router.Middlewares) {
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger)

	auth.Get("/settings", h.handleSettingsPage)
	auth.Post("/settings/davSession", h.createDavSession)
}

func (h *settingsHandler) String() string {
	return "web.settings"
}

func (h *settingsHandler) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	webSessions, err := h.webSessions.GetUserSessions(ctx, currentSession.UserID())
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the websessions: %w", err))
		return
	}

	davSessions, err := h.davSessions.GetAllForUser(ctx, currentSession.UserID(), &storage.PaginateCmd{Limit: 10})
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the davsessions: %w", err))
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "settings/index.tmpl", map[string]interface{}{
		"currentSession": currentSession,
		"webSessions":    webSessions,
		"davSessions":    davSessions,
		"oauthSessions":  []string{},
	})
}

func (h *settingsHandler) createDavSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	user, err := h.users.GetByID(ctx, currentSession.UserID())
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	session, secret, err := h.davSessions.Create(ctx, &davsessions.CreateCmd{
		UserID: currentSession.UserID(),
		FSRoot: user.RootFS(),
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "settings/show-dav-credentials.tmpl", map[string]interface{}{
		"session": session,
		"secret":  secret,
	})
}
