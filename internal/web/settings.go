package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type renderUsersCmd struct {
	User    *users.User
	Session *websessions.Session
	Error   error
}

type settingsHandler struct {
	html        html.Writer
	webSessions websessions.Service
	davSessions davsessions.Service
	spaces      spaces.Service
	users       users.Service
	uuid        uuid.Service
	auth        *Authenticator
}

func newSettingsHandler(
	tools tools.Tools,
	html html.Writer,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	spaces spaces.Service,
	users users.Service,
	authent *Authenticator,
) *settingsHandler {
	return &settingsHandler{
		html:        html,
		webSessions: webSessions,
		davSessions: davSessions,
		spaces:      spaces,
		users:       users,
		uuid:        tools.UUID(),
		auth:        authent,
	}
}

func (h *settingsHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/settings", http.RedirectHandler("/settings/general", http.StatusMovedPermanently).ServeHTTP)

	r.Get("/settings/security", h.getSecurityPage)
	r.Get("/settings/security/webdav", h.getWebDAVForm)
	r.Post("/settings/security/webdav", h.createDavSession)
	r.Post("/settings/security/webdav/{sessionID}/delete", h.deleteDavSession)
	r.Post("/settings/security/browsers/{sessionToken}/delete", h.deleteWebSession)

	r.Get("/settings/general", h.getGeneralPage)

	r.Get("/settings/users", h.getUsers)
	r.Post("/settings/users", h.createUser)
	r.Post("/settings/users/{userID}/delete", h.deleteUser)
}

func (h *settingsHandler) String() string {
	return "web.settings"
}

func (h *settingsHandler) getSecurityPage(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.renderSecurityPage(w, r, &securityCmd{
		User:    user,
		Session: session,
	})
}

type securityCmd struct {
	User    *users.User
	Session *websessions.Session
}

func (h *settingsHandler) renderSecurityPage(w http.ResponseWriter, r *http.Request, cmd *securityCmd) {
	ctx := r.Context()

	webSessions, err := h.webSessions.GetAllForUser(ctx, cmd.User.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to webSessions.GetAllForUser: %w", err))
		return
	}

	davSessions, err := h.davSessions.GetAllForUser(ctx, cmd.User.ID(), &storage.PaginateCmd{Limit: 20})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to davSessions.GetAllForUser: %w", err))
		return
	}

	spaceList, err := h.spaces.GetAllUserSpaces(ctx, cmd.User.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to spaces.GetAllForUser: %w", err))
		return
	}

	spacesMap := make(map[uuid.UUID]spaces.Space)
	for _, space := range spaceList {
		spacesMap[space.ID()] = space
	}

	h.html.WriteHTML(w, r, http.StatusOK, "settings/security/content.tmpl", map[string]interface{}{
		"isAdmin":        cmd.User.IsAdmin(),
		"currentSession": cmd.Session,
		"webSessions":    webSessions,
		"devices":        davSessions,
		"spaces":         spacesMap,
	})
}

func (h *settingsHandler) getWebDAVForm(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.renderWebDAVForm(w, r, &webdavFormCmd{Error: nil, User: user})
}

type webdavFormCmd struct {
	Error error
	User  *users.User
}

func (h *settingsHandler) renderWebDAVForm(w http.ResponseWriter, r *http.Request, cmd *webdavFormCmd) {
	spaces, err := h.spaces.GetAllUserSpaces(r.Context(), cmd.User.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllUserSpaces: %w", err))
		return
	}

	status := http.StatusOK
	if cmd.Error != nil {
		status = http.StatusUnprocessableEntity
	}

	h.html.WriteHTML(w, r, status, "settings/security/webdav-form.tmpl", map[string]interface{}{
		"error":  cmd.Error,
		"spaces": spaces,
	})
}

func (h *settingsHandler) getGeneralPage(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.renderGeneralPage(w, r)
}

func (h *settingsHandler) renderGeneralPage(w http.ResponseWriter, r *http.Request) {
	h.html.WriteHTML(w, r, http.StatusOK, "settings/general/content.tmpl", map[string]interface{}{})
}

func (h *settingsHandler) createDavSession(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	spaceID, err := h.uuid.Parse(r.FormValue("space"))
	if err != nil {
		h.renderWebDAVForm(w, r, &webdavFormCmd{User: user, Error: errors.New("invalid space id")})
		return
	}

	newSession, secret, err := h.davSessions.Create(r.Context(), &davsessions.CreateCmd{
		UserID:   user.ID(),
		Name:     r.FormValue("name"),
		Username: user.Username(),
		SpaceID:  spaceID,
	})
	if errors.Is(err, errs.ErrValidation) {
		h.renderWebDAVForm(w, r, &webdavFormCmd{User: user, Error: err})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Create dav session: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusCreated, "settings/security/webdav-result.tmpl", map[string]interface{}{
		"secret":     secret,
		"newSession": newSession,
	})
}

func (h *settingsHandler) deleteDavSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	sessionID, err := h.uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("invalid session id in dav session deletion: %w", err))
		return
	}

	err = h.davSessions.Delete(r.Context(), &davsessions.DeleteCmd{
		UserID:    user.ID(),
		SessionID: sessionID,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to davSession.Delete: %w", err))
		return
	}

	h.renderSecurityPage(w, r, &securityCmd{User: user, Session: session})
}

func (h *settingsHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AdminOnly)
	if abort {
		return
	}

	h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: nil})
}

func (h *settingsHandler) renderUsers(w http.ResponseWriter, r *http.Request, cmd renderUsersCmd) {
	ctx := r.Context()

	users, err := h.users.GetAll(ctx, &storage.PaginateCmd{
		StartAfter: map[string]string{"username": ""},
		Limit:      10,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to users.GetAll: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "settings/users.tmpl", map[string]interface{}{
		"isAdmin": cmd.User.IsAdmin(),
		"current": cmd.User,
		"users":   users,
		"error":   cmd.Error,
	})
}

func (h *settingsHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AdminOnly)
	if abort {
		return
	}

	userToDelete, err := h.uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: errors.New("invalid user id")})
		return
	}

	err = h.users.AddToDeletion(r.Context(), userToDelete)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to users.AddToDeletion: %w", err))
		return
	}

	h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: nil})
}

func (h *settingsHandler) createUser(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AdminOnly)
	if abort {
		return
	}

	_, err := h.users.Create(r.Context(), &users.CreateCmd{
		Username: r.FormValue("username"),
		Password: secret.NewText(r.FormValue("password")),
		IsAdmin:  r.FormValue("role") == "admin",
	})
	if errors.Is(err, errs.ErrValidation) {
		h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: err})
		return
	}
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to users.Create: %w", err))
		return
	}

	h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: nil})
}

func (h *settingsHandler) deleteWebSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AdminOnly)
	if abort {
		return
	}

	sessionID, err := h.uuid.Parse(chi.URLParam(r, "sessionToken"))
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("invalid sessionToken for the browser deletion: %w", err))
		return
	}

	err = h.webSessions.Delete(r.Context(), &websessions.DeleteCmd{
		UserID: user.ID(),
		Token:  secret.NewText(string(sessionID)),
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to delete a session: %w", err))
		return
	}

	h.renderSecurityPage(w, r, &securityCmd{User: user, Session: session})
}
