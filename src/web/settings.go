package web

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/myminicloud/myminicloud/src/service/websessions"
	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/response"
	"github.com/myminicloud/myminicloud/src/tools/router"
)

type settingsHandler struct {
	response    response.Writer
	webSessions websessions.Service
}

func newSettingsHandler(tools tools.Tools, webSessions websessions.Service) *settingsHandler {
	return &settingsHandler{
		response:    tools.ResWriter(),
		webSessions: webSessions,
	}
}

func (h *settingsHandler) Register(r chi.Router, mids router.Middlewares) {
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger)

	auth.HandleFunc("/settings", h.handleSettingsPage)
}

func (h *settingsHandler) String() string {
	return "web.settings"
}

func (h *settingsHandler) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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

	h.response.WriteHTML(w, http.StatusOK, "settings/index.tmpl", map[string]interface{}{
		"currentSession": currentSession,
		"webSessions":    webSessions,
	})
}
