package web

import (
	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthconsents"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/service/websessions"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
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
