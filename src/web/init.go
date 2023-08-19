package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/router"
)

type HTTPHandler struct {
	auth     *authHandler
	settings *settingsHandler
}

func NewHTTPHandler(
	tools tools.Tools,
	users users.Service,
	clients oauthclients.Service,
	oauthConsent oauthconsents.Service,
	webSessions websessions.Service,
) *HTTPHandler {
	return &HTTPHandler{
		auth:     newAuthHandler(tools, users, clients, oauthConsent, webSessions),
		settings: newSettingsHandler(tools, webSessions),
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	h.auth.Register(r, mids)
	h.settings.Register(r, mids)
}

func (h *HTTPHandler) String() string {
	return "web"
}
