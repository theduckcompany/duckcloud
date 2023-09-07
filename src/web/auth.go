package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/oauth2"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type authHandler struct {
	response response.Writer
	uuid     uuid.Service

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
		uuid:         tools.UUID(),
		response:     tools.ResWriter(),
		users:        users,
		clients:      clients,
		oauthConsent: oauthConsent,
		webSession:   webSessions,
	}
}

func (h *authHandler) Register(r chi.Router, mids router.Middlewares) {
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger, mids.CORS)

	auth.Get("/login", h.printLoginPage)
	auth.Post("/login", h.applyLogin)
	auth.HandleFunc("/consent", h.handleConsentPage)
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
	h.response.WriteHTML(w, http.StatusOK, "auth/login.tmpl", true, nil)
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
		h.response.WriteHTML(w, status, "auth/login.tmpl", true, map[string]interface{}{
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
		h.response.WriteJSONError(w, fmt.Errorf("failed to create the websession: %w", err))
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
			h.response.WriteJSONError(w, errs.BadRequest(oauth2.ErrClientNotFound, "client not found"))
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

	clientID, err := h.uuid.Parse(r.FormValue("client_id"))
	if err != nil {
		h.response.WriteJSONError(w, errs.BadRequest(oauth2.ErrClientNotFound, "client not found"))
		return
	}

	client, err := h.clients.GetByID(r.Context(), clientID)
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

	h.response.WriteHTML(w, http.StatusOK, "auth/consent.tmpl", true, map[string]interface{}{
		"clientName": client.Name(),
		"username":   user.Username,
		"scope":      strings.Split(r.FormValue("scope"), ","),
		"redirect":   template.URL("/consent?" + r.Form.Encode()),
	})
}
