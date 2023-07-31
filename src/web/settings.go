package web

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
)

type settingsHandler struct {
	response response.Writer
}

func newSettingsHandler(tools tools.Tools) *settingsHandler {
	return &settingsHandler{
		response: tools.ResWriter(),
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
	if r.Method == http.MethodGet {
		h.response.WriteHTML(w, http.StatusOK, "settings/index.tmpl", nil)
		return
	}
}
