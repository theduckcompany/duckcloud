package settings

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
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/settings/security"
)

type securityCmd struct {
	User    *users.User
	Session *websessions.Session
}

type webdavFormCmd struct {
	Error error
	User  *users.User
}

type securityPage struct {
	auth        *auth.Authenticator
	webSessions websessions.Service
	html        html.Writer
	davSessions davsessions.Service
	spaces      spaces.Service
	uuid        uuid.Service
	users       users.Service
}

func newSecurityPage(
	tools tools.Tools,
	html html.Writer,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	spaces spaces.Service,
	users users.Service,
	authent *auth.Authenticator,
) *securityPage {
	return &securityPage{
		auth:        authent,
		webSessions: webSessions,
		html:        html,
		davSessions: davSessions,
		spaces:      spaces,
		uuid:        tools.UUID(),
		users:       users,
	}
}

func (h *securityPage) Register(r chi.Router, mids *router.Middlewares) {
	r.Get("/settings/security", h.getSecurityPage)
	r.Get("/settings/security/webdav", h.getWebDAVForm)
	r.Post("/settings/security/webdav", h.createDavSession)
	r.Post("/settings/security/webdav/{sessionID}/delete", h.deleteDavSession)
	r.Post("/settings/security/browsers/{sessionToken}/delete", h.deleteWebSession)
	r.Get("/settings/security/password", h.getPasswordForm)
	r.Post("/settings/security/password", h.updatePassword)
}

func (h *securityPage) getSecurityPage(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	h.renderSecurityPage(w, r, &securityCmd{
		User:    user,
		Session: session,
	})
}

func (h *securityPage) getWebDAVForm(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	h.renderWebDAVForm(w, r, &webdavFormCmd{Error: nil, User: user})
}

func (h *securityPage) createDavSession(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
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

	h.html.WriteHTMLTemplate(w, r, http.StatusCreated, &security.WebdavResultTemplate{
		Secret:     secret, // TODO: Use secrets.String
		NewSession: newSession,
	})
}

func (h *securityPage) deleteDavSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
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

func (h *securityPage) deleteWebSession(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
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

func (h *securityPage) getPasswordForm(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	h.renderPasswordForm(w, r, &passwordFormCmd{Error: nil})
}

func (h *securityPage) updatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	currentPassword := secret.NewText(r.FormValue("current"))

	_, err := h.users.Authenticate(ctx, user.Username(), currentPassword)
	if errors.Is(err, users.ErrInvalidPassword) {
		h.renderPasswordForm(w, r, &passwordFormCmd{Error: errors.New("invalid current password")})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Authenticate the user: %w", err))
		return
	}

	newPassword := secret.NewText(r.FormValue("new"))
	confirmPassword := secret.NewText(r.FormValue("confirm"))

	if !confirmPassword.Equals(newPassword) {
		h.renderPasswordForm(w, r, &passwordFormCmd{Error: errors.New("the new password and the confirmation are different")})
		return
	}

	err = h.users.UpdateUserPassword(ctx, &users.UpdatePasswordCmd{
		UserID:      user.ID(),
		NewPassword: newPassword,
	})
	if errors.Is(err, errs.ErrValidation) {
		h.renderPasswordForm(w, r, &passwordFormCmd{Error: err})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to UpdateUserPassword: %w", err))
		return
	}

	h.renderSecurityPage(w, r, &securityCmd{User: user, Session: session})
}

func (h *securityPage) renderPasswordForm(w http.ResponseWriter, r *http.Request, cmd *passwordFormCmd) {
	status := http.StatusOK

	var errStr string
	if cmd.Error != nil {
		status = http.StatusUnprocessableEntity
		errStr = cmd.Error.Error()
	}

	h.html.WriteHTMLTemplate(w, r, status, &security.PasswordFormTemplate{
		Error: errStr,
	})
}

func (h *securityPage) renderSecurityPage(w http.ResponseWriter, r *http.Request, cmd *securityCmd) {
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

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &security.ContentTemplate{
		IsAdmin:        cmd.User.IsAdmin(),
		CurrentSession: cmd.Session,
		WebSessions:    webSessions,
		Devices:        davSessions,
		Spaces:         spacesMap,
	})
}

func (h *securityPage) renderWebDAVForm(w http.ResponseWriter, r *http.Request, cmd *webdavFormCmd) {
	spaces, err := h.spaces.GetAllUserSpaces(r.Context(), cmd.User.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllUserSpaces: %w", err))
		return
	}

	status := http.StatusOK
	if cmd.Error != nil {
		status = http.StatusUnprocessableEntity
	}

	h.html.WriteHTMLTemplate(w, r, status, &security.WebdavFormTemplate{
		Error:  cmd.Error,
		Spaces: spaces,
	})
}
