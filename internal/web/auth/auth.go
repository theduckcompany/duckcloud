package auth

import (
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type Handler struct {
	html html.Writer
	uuid uuid.Service

	auth         *Authenticator
	users        users.Service
	clients      oauthclients.Service
	webSession   websessions.Service
	oauthConsent oauthconsents.Service
	loginPage    *loginPage
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
		loginPage:    newLoginPage(htmlWriter, webSessions, users, clients, tools),
		uuid:         tools.UUID(),
		auth:         auth,
		html:         htmlWriter,
		users:        users,
		clients:      clients,
		oauthConsent: oauthConsent,
		webSession:   webSessions,
	}
}

func (h *Handler) Register(r chi.Router, mids *router.Middlewares) {
	h.loginPage.Register(r, mids)

	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.HandleFunc("/consent", h.handleConsentPage)
}

func (h *Handler) handleConsentPage(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	user, session, abort := h.auth.GetUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	clientID, err := h.uuid.Parse(r.FormValue("client_id"))
	if err != nil {
		h.printClientErrorPage(w, r, errors.New("invalid client_id"))
		return
	}

	client, err := h.clients.GetByID(r.Context(), clientID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, err)
		return
	}
	if client == nil {
		h.printClientErrorPage(w, r, errors.New("invalid client_id"))
		return
	}

	if r.Method == http.MethodPost {
		consent, err := h.oauthConsent.Create(r.Context(), &oauthconsents.CreateCmd{
			UserID:       user.ID(),
			SessionToken: session.Token().Raw(),
			ClientID:     client.GetID(),
			Scopes:       strings.Split(r.FormValue("scope"), ","),
		})
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, err)
			return
		}

		r.Form.Add("consent_id", string(consent.ID()))
		http.Redirect(w, r, "/auth/authorize?"+r.Form.Encode(), http.StatusFound)
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "auth/consent", map[string]interface{}{
		"clientName": client.Name(),
		"username":   user.Username,
		"scope":      strings.Split(r.FormValue("scope"), ","),
		"redirect":   template.URL("/consent?" + r.Form.Encode()),
	})
}

func (h *Handler) printClientErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	reqID, ok := r.Context().Value(middleware.RequestIDKey).(string)
	if !ok {
		reqID = "??1?"
	}

	h.html.WriteHTML(w, r, http.StatusBadRequest, "auth/error", map[string]interface{}{
		"reqID": reqID,
		"error": err,
	})
}
