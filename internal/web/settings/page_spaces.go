package settings

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	spacestmpl "github.com/theduckcompany/duckcloud/internal/web/html/templates/settings/spaces"
)

type SpacesPage struct {
	html      html.Writer
	spaces    spaces.Service
	users     users.Service
	scheduler scheduler.Service
	auth      *auth.Authenticator
	uuid      uuid.Service
}

func NewSpacesPage(
	html html.Writer,
	spaces spaces.Service,
	users users.Service,
	authent *auth.Authenticator,
	scheduler scheduler.Service,
	tools tools.Tools,
) *SpacesPage {
	return &SpacesPage{
		html:      html,
		spaces:    spaces,
		users:     users,
		scheduler: scheduler,
		auth:      authent,
		uuid:      tools.UUID(),
	}
}

func (h *SpacesPage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.Defaults()...)
	}
	r.Get("/settings/spaces", h.getContent)
	r.Get("/settings/spaces/new", h.getCreateSpaceModal)
	r.Post("/settings/spaces/create", h.createSpace)
	r.Post("/settings/spaces/{spaceID}/delete", h.deleteSpace)
}

func (h *SpacesPage) getContent(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	if !user.IsAdmin() {
		logger.LogEntrySetError(r.Context(), fmt.Errorf("%q is not an admin", user.Username()))
		http.Redirect(w, r, "/settings", http.StatusUnauthorized)
		return
	}

	h.renderContent(w, r, user)
}

func (h *SpacesPage) renderContent(w http.ResponseWriter, r *http.Request, user *users.User) {
	usersArray, err := h.users.GetAll(r.Context(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllUsers: %w", err))
		return
	}

	usersMap := make(map[uuid.UUID]users.User, len(usersArray))
	for _, u := range usersArray {
		usersMap[u.ID()] = u
	}

	spaces, err := h.spaces.GetAllSpaces(r.Context(), user, nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetAllSpaces: %w", err))
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &spacestmpl.ContentTemplate{
		IsAdmin: user.IsAdmin(),
		Spaces:  spaces,
		Users:   usersMap,
	})
}

func (h *SpacesPage) deleteSpace(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
	if abort {
		return
	}

	spaceID, err := h.uuid.Parse(chi.URLParam(r, "spaceID"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		logger.LogEntrySetError(r.Context(), fmt.Errorf("spaceID %q not found", chi.URLParam(r, "spaceID")))
		return
	}

	err = h.spaces.Delete(r.Context(), user, spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Delete the space: %w", err))
		return
	}

	h.renderContent(w, r, user)
}

func (h *SpacesPage) getCreateSpaceModal(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
	if abort {
		return
	}

	h.renderCreateSpaceModal(w, r, user)
}

func (h *SpacesPage) createSpace(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AdminOnly)
	if abort {
		return
	}

	r.ParseForm()

	var err error
	owners := make([]uuid.UUID, len(r.Form["selectedUsers"]))

	for i, idStr := range r.Form["selectedUsers"] {
		owners[i], err = h.uuid.Parse(idStr)
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("invalid user id (%q): %w", idStr, err))
			return
		}
	}

	err = h.scheduler.RegisterSpaceCreateTask(r.Context(), &scheduler.SpaceCreateArgs{
		UserID: user.ID(),
		Name:   r.Form.Get("name"),
		Owners: owners,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to create the space: %w", err))
		return
	}

	h.renderContent(w, r, user)
}

func (h *SpacesPage) renderCreateSpaceModal(w http.ResponseWriter, r *http.Request, user *users.User) {
	allUsers, err := h.users.GetAll(r.Context(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Get all the users: %w", err))
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &spacestmpl.CreateSpaceModal{
		IsAdmin: user.IsAdmin(),
		Selection: spacestmpl.UserSelectionTemplate{
			UnselectedUsers: allUsers,
			SelectedUsers:   []users.User{*user},
		},
	})
}
