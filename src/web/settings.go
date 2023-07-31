package web

import (
	"fmt"
	"net/http"

	"github.com/Peltoche/neurone/src/service/websessions"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
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

	if r.Method == http.MethodGet {
		currentSession, err := h.webSessions.GetFromReq(r)
		if err != nil {
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusFound)
		}

		webSessions, err := h.webSessions.GetUserSessions(ctx, currentSession.UserID)
		if err != nil {
			h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the websessions: %w", err))
			return
		}

		h.response.WriteHTML(w, http.StatusOK, "settings/index.tmpl", map[string]interface{}{
			"session":     currentSession,
			"websessions": webSessions,
		})

		return
	}
}
