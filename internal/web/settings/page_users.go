package settings

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type usersPage struct {
	html  html.Writer
	users users.Service
	auth  *auth.Authenticator
	uuid  uuid.Service
}

func newUsersPage(
	tools tools.Tools,
	html html.Writer,
	users users.Service,
	authent *auth.Authenticator,
) *usersPage {
	return &usersPage{
		html:  html,
		users: users,
		auth:  authent,
		uuid:  tools.UUID(),
	}
}

func (h *usersPage) Register(r chi.Router, mids *router.Middlewares) {
	r.Get("/settings/users", h.getUsers)
	r.Post("/settings/users", h.createUser)
	r.Get("/settings/users/new", h.getUsersRegistrationForm)
	r.Post("/settings/users/{userID}/delete", h.deleteUser)
}

func (h *usersPage) getUsers(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
	if abort {
		return
	}

	h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: nil})
}

func (h *usersPage) createUser(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
	if abort {
		return
	}

	_, err := h.users.Create(r.Context(), &users.CreateCmd{
		User:     user,
		Username: r.FormValue("username"),
		Password: secret.NewText(r.FormValue("password")),
		IsAdmin:  r.FormValue("role") == "admin",
	})
	if errors.Is(err, errs.ErrValidation) {
		h.renderUsersRegistrationForm(w, r, err)
		return
	}
	if errors.Is(err, users.ErrUsernameTaken) {
		h.renderUsersRegistrationForm(w, r, errors.New("username already taken"))
		return
	}
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to users.Create: %w", err))
		return
	}

	h.renderUsers(w, r, renderUsersCmd{User: user, Session: session, Error: nil})
}

func (h *usersPage) getUsersRegistrationForm(w http.ResponseWriter, r *http.Request) {
	h.renderUsersRegistrationForm(w, r, nil)
}

func (h *usersPage) deleteUser(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
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

func (h *usersPage) renderUsersRegistrationForm(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusOK
	if err != nil {
		status = http.StatusUnprocessableEntity
	}

	h.html.WriteHTML(w, r, status, "settings/users/registration-form.tmpl", map[string]interface{}{
		"error": err,
	})
}

func (h *usersPage) renderUsers(w http.ResponseWriter, r *http.Request, cmd renderUsersCmd) {
	ctx := r.Context()

	users, err := h.users.GetAll(ctx, &storage.PaginateCmd{
		StartAfter: map[string]string{"username": ""},
		Limit:      20,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to users.GetAll: %w", err))
		return
	}

	status := http.StatusOK
	if cmd.Error != nil {
		status = http.StatusUnprocessableEntity
	}

	h.html.WriteHTML(w, r, status, "settings/users/content.tmpl", map[string]interface{}{
		"isAdmin": cmd.User.IsAdmin(),
		"current": cmd.User,
		"users":   users,
		"error":   cmd.Error,
	})
}
