package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/browser"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/settings"
)

type Config struct {
	HTML html.Config
}

type HTTPHandler struct {
	auth     *auth.Handler
	browser  *browser.Handler
	settings *settings.Handler
	home     *homeHandler
}

func NewHTTPHandler(
	cfg Config,
	tools tools.Tools,
	users users.Service,
	clients oauthclients.Service,
	files files.Service,
	oauthConsent oauthconsents.Service,
	webSessions websessions.Service,
	spaces spaces.Service,
	davSessions davsessions.Service,
	fs dfs.Service,
) *HTTPHandler {
	htmlRenderer := html.NewRenderer(cfg.HTML)
	authenticator := auth.NewAuthenticator(webSessions, users, htmlRenderer)

	return &HTTPHandler{
		auth:     auth.NewHandler(tools, htmlRenderer, authenticator, users, clients, oauthConsent, webSessions),
		browser:  browser.NewHandler(tools, htmlRenderer, spaces, files, authenticator, fs),
		settings: settings.NewHandler(tools, htmlRenderer, webSessions, davSessions, spaces, users, authenticator),
		home:     newHomeHandler(htmlRenderer, authenticator),
	}
}

func (h *HTTPHandler) Register(r chi.Router, mids *router.Middlewares) {
	h.auth.Register(r, mids)
	h.browser.Register(r, mids)

	h.settings.Register(r, mids)
	h.home.Register(r, mids)
}

func (h *HTTPHandler) String() string {
	return "web"
}
