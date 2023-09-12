package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/web/html"
)

type Config struct {
	HTML html.Config `json:"html"`
}

type HTTPHandler struct {
	auth     *authHandler
	settings *settingsHandler
	home     *homeHandler
	browser  *browserHandler
}

func NewHTTPHandler(
	cfg Config,
	tools tools.Tools,
	users users.Service,
	clients oauthclients.Service,
	oauthConsent oauthconsents.Service,
	webSessions websessions.Service,
	folders folders.Service,
	files files.Service,
	davSessions davsessions.Service,
	inodes inodes.Service,
) *HTTPHandler {
	htmlRenderer := html.NewRenderer(cfg.HTML)
	auth := NewAuthenticator(webSessions, users, htmlRenderer)

	return &HTTPHandler{
		auth:     newAuthHandler(tools, htmlRenderer, auth, users, clients, oauthConsent, webSessions),
		settings: newSettingsHandler(tools, htmlRenderer, webSessions, davSessions, folders, users, auth),
		home:     newHomeHandler(htmlRenderer, auth),
		browser:  newBrowserHandler(tools, htmlRenderer, folders, inodes, files, tools.UUID(), auth),
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids *router.Middlewares) {
	h.auth.Register(r, mids)
	h.settings.Register(r, mids)
	h.home.Register(r, mids)
	h.browser.Register(r, mids)
}

func (h *HTTPHandler) String() string {
	return "web"
}
