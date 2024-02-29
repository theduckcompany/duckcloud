package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

type LoginPage struct {
	webSessions websessions.Service
	uuid        uuid.Service
	html        html.Writer
	users       users.Service
	clients     oauthclients.Service
}

func NewLoginPage(
	html html.Writer,
	webSessions websessions.Service,
	users users.Service,
	clients oauthclients.Service,
	tools tools.Tools,
) *LoginPage {
	return &LoginPage{
		html:        html,
		webSessions: webSessions,
		users:       users,
		clients:     clients,
		uuid:        tools.UUID(),
	}
}

func (h *LoginPage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.Defaults()...)
	}

	r.Get("/login", h.printPage)
	r.Post("/login", h.applyLogin)
}

func (h *LoginPage) printPage(w http.ResponseWriter, r *http.Request) {
	currentSession, _ := h.webSessions.GetFromReq(r)

	if currentSession != nil {
		h.chooseRedirection(w, r)
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.LoginPageTmpl{})
}

func (h *LoginPage) applyLogin(w http.ResponseWriter, r *http.Request) {
	tmpl := auth.LoginPageTmpl{}

	tmpl.UsernameContent = r.FormValue("username")

	user, err := h.users.Authenticate(r.Context(), r.FormValue("username"), secret.NewText(r.FormValue("password")))
	var status int
	switch {
	case err == nil:
		// continue
	case errors.Is(err, users.ErrInvalidUsername):
		tmpl.UsernameError = "User doesn't exists"
		status = http.StatusBadRequest
	case errors.Is(err, users.ErrInvalidPassword):
		tmpl.PasswordError = "Invalid password"
		status = http.StatusBadRequest
	default:
		h.html.WriteHTMLErrorPage(w, r, err)
		return
	}

	if err != nil {
		h.html.WriteHTMLTemplate(w, r, status, &tmpl)
		return
	}

	session, err := h.webSessions.Create(r.Context(), &websessions.CreateCmd{
		UserID:     user.ID(),
		UserAgent:  r.Header.Get("User-Agent"),
		RemoteAddr: r.RemoteAddr,
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

func (h *LoginPage) chooseRedirection(w http.ResponseWriter, r *http.Request) {
	var client *oauthclients.Client
	clientID, err := h.uuid.Parse(r.FormValue("client_id"))
	if err == nil {
		client, err = h.clients.GetByID(r.Context(), clientID)
		if err != nil {
			reqID, ok := r.Context().Value(middleware.RequestIDKey).(string)
			if !ok {
				reqID = "????"
			}

			h.html.WriteHTMLTemplate(w, r, http.StatusBadRequest, &auth.ErrorPageTmpl{
				ErrorMsg:  "Oauth client not found",
				RequestID: reqID,
			})
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
