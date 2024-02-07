package auth

import (
	"errors"
	"fmt"
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
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
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
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/login", h.printLoginPage)
	r.Post("/login", h.applyLogin)
	r.HandleFunc("/consent", h.handleConsentPage)
}

func (h *Handler) printLoginPage(w http.ResponseWriter, r *http.Request) {
	currentSession, _ := h.webSession.GetFromReq(r)

	if currentSession != nil {
		h.chooseRedirection(w, r)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	h.html.WriteHTML(w, r, http.StatusOK, "auth/page", nil)
}

func (h *Handler) applyLogin(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	inputs := map[string]string{}
	loginErrors := map[string]string{}

	inputs["username"] = r.FormValue("username")

	user, err := h.users.Authenticate(r.Context(), r.FormValue("username"), secret.NewText(r.FormValue("password")))
	var status int
	switch {
	case err == nil:
		// continue
	case errors.Is(err, users.ErrInvalidUsername):
		loginErrors["username"] = "User doesn't exists"
		status = http.StatusBadRequest
	case errors.Is(err, users.ErrInvalidPassword):
		loginErrors["password"] = "Invalid password"
		status = http.StatusBadRequest
	default:
		loginErrors["notif"] = "Unexpected error"
		status = http.StatusBadRequest
	}

	if len(loginErrors) > 0 {
		h.html.WriteHTML(w, r, status, "auth/page", map[string]interface{}{
			"inputs": inputs,
			"errors": loginErrors,
		})
		return
	}

	session, err := h.webSession.Create(r.Context(), &websessions.CreateCmd{
		UserID: string(user.ID()),
		Req:    r,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to create the websession: %w", err))
		return
	}

	// TODO: Handle the expiration time with the "Remember me" option
	c := http.Cookie{
		Name:     "session_token",
		Value:    session.Token().Raw(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, &c)

	h.chooseRedirection(w, r)
}

func (h *Handler) chooseRedirection(w http.ResponseWriter, r *http.Request) {
	var client *oauthclients.Client
	clientID, err := h.uuid.Parse(r.FormValue("client_id"))
	if err == nil {
		client, err = h.clients.GetByID(r.Context(), clientID)
		if err != nil {
			h.printClientErrorPage(w, r, errors.New("Oauth client not found"))
			return
		}
	}

	switch {
	case client == nil:
		http.Redirect(w, r, "/", http.StatusFound)
	case client.SkipValidation():
		http.Redirect(w, r, "/auth/authorize", http.StatusFound)
	default:
		http.Redirect(w, r, "/consent?"+r.Form.Encode(), http.StatusFound)
	}
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
