package web

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
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type authHandler struct {
	html html.Writer
	uuid uuid.Service

	auth         *Authenticator
	users        users.Service
	clients      oauthclients.Service
	webSession   websessions.Service
	oauthConsent oauthconsents.Service
}

func newAuthHandler(
	tools tools.Tools,
	htmlWriter html.Writer,
	auth *Authenticator,
	users users.Service,
	clients oauthclients.Service,
	oauthConsent oauthconsents.Service,
	webSessions websessions.Service,
) *authHandler {
	return &authHandler{
		uuid:         tools.UUID(),
		auth:         auth,
		html:         htmlWriter,
		users:        users,
		clients:      clients,
		oauthConsent: oauthConsent,
		webSession:   webSessions,
	}
}

func (h *authHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/login", h.printLoginPage)
	r.Post("/login", h.applyLogin)
	r.HandleFunc("/consent", h.handleConsentPage)
}

func (h *authHandler) String() string {
	return "web.auth"
}

func (h *authHandler) printLoginPage(w http.ResponseWriter, r *http.Request) {
	currentSession, _ := h.webSession.GetFromReq(r)

	if currentSession != nil {
		h.chooseRedirection(w, r)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	h.html.WriteHTML(w, r, http.StatusOK, "auth/login.tmpl", nil)
}

func (h *authHandler) applyLogin(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	inputs := map[string]string{}
	loginErrors := map[string]string{}

	inputs["username"] = r.FormValue("username")

	user, err := h.users.Authenticate(r.Context(), r.FormValue("username"), r.FormValue("password"))
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
		h.html.WriteHTML(w, r, status, "auth/login.tmpl", map[string]interface{}{
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
		Value:    session.Token(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, &c)

	h.chooseRedirection(w, r)
}

func (h *authHandler) chooseRedirection(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
	case client.SkipValidation():
		w.Header().Set("Location", "/auth/authorize")
		w.WriteHeader(http.StatusFound)
	default:
		w.Header().Set("Location", "/consent?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
	}
}

func (h *authHandler) handleConsentPage(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
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
			SessionToken: session.Token(),
			ClientID:     client.GetID(),
			Scopes:       strings.Split(r.FormValue("scope"), ","),
		})
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, err)
			return
		}

		r.Form.Add("consent_id", string(consent.ID()))
		w.Header().Set("Location", "/auth/authorize?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "auth/consent.tmpl", map[string]interface{}{
		"clientName": client.Name(),
		"username":   user.Username,
		"scope":      strings.Split(r.FormValue("scope"), ","),
		"redirect":   template.URL("/consent?" + r.Form.Encode()),
	})
}

func (h *authHandler) printClientErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	reqID, ok := r.Context().Value(middleware.RequestIDKey).(string)
	if !ok {
		reqID = "??1?"
	}

	h.html.WriteHTML(w, r, http.StatusBadRequest, "auth/error.tmpl", map[string]interface{}{
		"reqID": reqID,
		"error": err,
	})
}
