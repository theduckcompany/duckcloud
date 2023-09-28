package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type Config struct {
	HTML html.Config
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
	davSessions davsessions.Service,
	fs dfs.Service,
) *HTTPHandler {
	htmlRenderer := html.NewRenderer(cfg.HTML)
	auth := NewAuthenticator(webSessions, users, htmlRenderer)

	return &HTTPHandler{
		auth:     newAuthHandler(tools, htmlRenderer, auth, users, clients, oauthConsent, webSessions),
		settings: newSettingsHandler(tools, htmlRenderer, webSessions, davSessions, folders, users, auth),
		home:     newHomeHandler(htmlRenderer, auth),
		browser:  newBrowserHandler(tools, htmlRenderer, folders, auth, fs),
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
