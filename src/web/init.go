package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/myminicloud/myminicloud/src/service/oauthclients"
	"github.com/myminicloud/myminicloud/src/service/oauthconsents"
	"github.com/myminicloud/myminicloud/src/service/users"
	"github.com/myminicloud/myminicloud/src/service/websessions"
	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/router"
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
