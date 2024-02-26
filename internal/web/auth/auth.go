package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type Handler struct {
	loginPage   *loginPage
	consentPage *consentPage
}

func NewHandler(
	tools tools.Tools,
	htmlWriter html.Writer,
	auth *Authenticator,
	users users.Service,
	clients oauthclients.Service,
	oauthConsent oauthconsents.Service,
	webSessions websessions.Service,
) *Handler {
	return &Handler{
		loginPage:   newLoginPage(htmlWriter, webSessions, users, clients, tools),
		consentPage: newConsentPage(htmlWriter, auth, clients, oauthConsent, tools),
	}
}

func (h *Handler) Register(r chi.Router, mids *router.Middlewares) {
	h.loginPage.Register(r, mids)
	h.consentPage.Register(r, mids)
}
