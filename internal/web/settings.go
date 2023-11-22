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

type renderDavCmd struct {
	User       *users.User
	Session    *websessions.Session
	NewSession *davsessions.DavSession
	Secret     string
	Error      error
}

type renderBrowsersCmd struct {
	User    *users.User
	Session *websessions.Session
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

	r.Get("/settings", h.getBrowsersSessions)

	r.Get("/settings/browsers", h.getBrowsersSessions)
	r.Post("/settings/browsers/{sessionToken}/delete", h.deleteWebSession)

	r.Get("/settings/webdav", h.getDavSessions)
	r.Get("/settings/webdav/form", h.getCreateWebdavForm)
	r.Post("/settings/webdav", h.createDavSession)
	r.Post("/settings/webdav/{sessionID}/delete", h.deleteDavSession)

	r.Get("/settings/users", h.getUsers)
	r.Post("/settings/users", h.createUser)
	r.Post("/settings/users/{userID}/delete", h.deleteUser)
}

func (h *settingsHandler) String() string {
	return "web.settings"
}

func (h *settingsHandler) getCreateWebdavForm(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.renderDavForm(w, r, user.ID(), nil)
}

func (h *settingsHandler) getBrowsersSessions(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.renderBrowsersSessions(w, r, renderBrowsersCmd{User: user, Session: session})
}

func (h *settingsHandler) renderBrowsersSessions(w http.ResponseWriter, r *http.Request, cmd renderBrowsersCmd) {
	webSessions, err := h.webSessions.GetAllForUser(r.Context(), cmd.Session.UserID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllForUser: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "settings/browsers.tmpl", map[string]interface{}{
		"isAdmin":        cmd.User.IsAdmin(),
		"currentSession": cmd.Session,
		"webSessions":    webSessions,
	})
}

func (h *settingsHandler) getDavSessions(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	h.renderDavSessions(w, r, renderDavCmd{User: user, Session: session, NewSession: nil, Secret: "", Error: nil})
}

func (h *settingsHandler) renderDavForm(w http.ResponseWriter, r *http.Request, userID uuid.UUID, userErr error) {
	spaces, err := h.spaces.GetAllUserSpaces(r.Context(), userID, nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllUserSpaces: %w", err))
		return
	}

	status := http.StatusOK
	if userErr != nil {
		status = http.StatusUnprocessableEntity
	}

	h.html.WriteHTML(w, r, status, "settings/webdav-modal.tmpl", map[string]interface{}{
		"spaces": spaces,
		"error":  userErr,
	})
}

func (h *settingsHandler) renderDavSessions(w http.ResponseWriter, r *http.Request, cmd renderDavCmd) {
	ctx := r.Context()

	davSessions, err := h.davSessions.GetAllForUser(ctx, cmd.User.ID(), &storage.PaginateCmd{Limit: 10})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllForUser: %w", err))
		return
	}

	spaces, err := h.spaces.GetAllUserSpaces(ctx, cmd.User.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllUserSpaces: %w", err))
		return
	}

	status := http.StatusOK
	if cmd.Error != nil {
		status = http.StatusUnprocessableEntity
	}
	if cmd.NewSession != nil {
		status = http.StatusCreated
	}

	h.html.WriteHTML(w, r, status, "settings/webdav.tmpl", map[string]interface{}{
		"isAdmin":     cmd.User.IsAdmin(),
		"newSession":  cmd.NewSession,
		"davSessions": davSessions,
		"spaces":      spaces,
		"secret":      cmd.Secret,
		"error":       cmd.Error,
	})
}

func (h *settingsHandler) createDavSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	spaceID, err := h.uuid.Parse(r.FormValue("space"))
	if err != nil {
		h.renderDavForm(w, r, user.ID(), err)
		return
	}

	newSession, secret, err := h.davSessions.Create(r.Context(), &davsessions.CreateCmd{
		UserID:   user.ID(),
		Name:     r.FormValue("name"),
		Username: user.Username(),
		SpaceID:  spaceID,
	})
	if errors.Is(err, errs.ErrValidation) {
		h.renderDavForm(w, r, user.ID(), err)
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Create dav session: %w", err))
		return
	}

	h.renderDavSessions(w, r, renderDavCmd{
		User:       user,
		Session:    session,
		NewSession: newSession,
		Secret:     secret,
		Error:      nil,
	})
}

func (h *settingsHandler) deleteWebSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	err := h.webSessions.Delete(r.Context(), &websessions.DeleteCmd{
		UserID: user.ID(),
		Token:  secret.NewText(chi.URLParam(r, "sessionToken")),
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to websession.Delete: %w", err))
		return
	}

	h.renderBrowsersSessions(w, r, renderBrowsersCmd{User: user, Session: session})
}

func (h *settingsHandler) deleteDavSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	sessionID, err := h.uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		h.renderDavSessions(w, r, renderDavCmd{User: user, Session: session, Error: errors.New("invalid session id")})
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

	h.renderDavSessions(w, r, renderDavCmd{User: user, Session: session, Error: nil})
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
