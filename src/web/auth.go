package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/myminicloud/myminicloud/src/service/oauth2"
	"github.com/myminicloud/myminicloud/src/service/oauthclients"
	"github.com/myminicloud/myminicloud/src/service/oauthconsents"
	"github.com/myminicloud/myminicloud/src/service/users"
	"github.com/myminicloud/myminicloud/src/service/websessions"
	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/errs"
	"github.com/myminicloud/myminicloud/src/tools/response"
	"github.com/myminicloud/myminicloud/src/tools/router"
)

type authHandler struct {
	response response.Writer

	users        users.Service
	clients      oauthclients.Service
	webSession   websessions.Service
	oauthConsent oauthconsents.Service
}

func newAuthHandler(
	tools tools.Tools,
	users users.Service,
	clients oauthclients.Service,
	oauthConsent oauthconsents.Service,
	webSessions websessions.Service,
) *authHandler {
	return &authHandler{
		response:     tools.ResWriter(),
		users:        users,
		clients:      clients,
		oauthConsent: oauthConsent,
		webSession:   webSessions,
	}
}

func (h *authHandler) Register(r chi.Router, mids router.Middlewares) {
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger, mids.CORS)

	auth.HandleFunc("/login", h.handleLoginPage)
	auth.HandleFunc("/forgot", h.handleForgotPage)
	auth.HandleFunc("/consent", h.handleConsentPage)
}

func (h *authHandler) String() string {
	return "web.auth"
}

func (h *authHandler) handleForgotPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.response.WriteHTML(w, http.StatusOK, "auth/forgot.tmpl", nil)
		return
	}

	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented yet!"))
}

func (h *authHandler) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	if r.Method == http.MethodGet {
		h.response.WriteHTML(w, http.StatusOK, "auth/login.tmpl", nil)
		return
	}

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
		h.response.WriteHTML(w, status, "auth/login.tmpl", map[string]interface{}{
			"inputs": inputs,
			"errors": loginErrors,
		})
		return
	}

	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil {
		h.response.WriteJSONError(w, errs.BadRequest(oauth2.ErrClientNotFound, "client not found"))
		return
	}

	isWebAuthentication := false

	if client == nil {
		isWebAuthentication = true
		client, err = h.clients.GetByID(r.Context(), oauthclients.WebAppClientID)
		if err != nil {
			h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the web app client: %w", err))
			return
		}

		if client == nil {
			h.response.WriteJSONError(w, errors.New("web client doesn't exists"))
			return
		}
	}

	session, err := h.webSession.Create(r.Context(), &websessions.CreateCmd{
		UserID:   string(user.ID()),
		ClientID: client.GetID(),
		Req:      r,
	})
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to create the websession: %w", err))
		return
	}

	// TODO: Handle the expiration time with the "Remember me" option
	c := http.Cookie{
		Name:     "session_token",
		Value:    session.Token(),
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, &c)

	switch {
	case isWebAuthentication:
		w.Header().Set("Location", client.RedirectURI())
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

	session, err := h.webSession.GetFromReq(r)
	if errors.Is(err, websessions.ErrMissingSessionToken) || errors.Is(err, websessions.ErrSessionNotFound) {
		w.Header().Set("Location", "/login?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
		return
	}

	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	user, err := h.users.GetByID(r.Context(), session.UserID())
	if err != nil || user == nil {
		h.response.WriteJSONError(w, errs.BadRequest(fmt.Errorf("failed to find the user %q: %w", session.UserID(), err), "user not found"))
		return
	}

	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil {
		h.response.WriteJSONError(w, errs.BadRequest(err, "invalid request"))
		return
	}
	if client == nil {
		h.response.WriteJSONError(w, errs.BadRequest(oauth2.ErrClientNotFound, "invalid request"))
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
			h.response.WriteJSONError(w, fmt.Errorf("failed to create the consent: %w", err))
			return
		}

		r.Form.Add("consent_id", string(consent.ID()))
		w.Header().Set("Location", "/auth/authorize?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "auth/consent.tmpl", map[string]interface{}{
		"clientName": client.Name(),
		"username":   user.Username,
		"scope":      strings.Split(r.FormValue("scope"), ","),
		"redirect":   template.URL("/consent?" + r.Form.Encode()),
	})
}
