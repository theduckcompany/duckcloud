package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type settingsHandler struct {
	response    response.Writer
	webSessions websessions.Service
	davSessions davsessions.Service
	folders     folders.Service
	users       users.Service
	uuid        uuid.Service
}

func newSettingsHandler(
	tools tools.Tools,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	folders folders.Service,
	users users.Service,
) *settingsHandler {
	return &settingsHandler{
		response:    tools.ResWriter(),
		webSessions: webSessions,
		davSessions: davSessions,
		folders:     folders,
		users:       users,
		uuid:        tools.UUID(),
	}
}

func (h *settingsHandler) Register(r chi.Router, mids router.Middlewares) {
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger)

	auth.Get("/settings", h.getBrowsersSessions)

	auth.Get("/settings/browsers", h.getBrowsersSessions)
	auth.Post("/settings/browsers/{sessionToken}/delete", h.deleteWebSession)

	auth.Get("/settings/webdav", h.getDavSessions)
	auth.Post("/settings/webdav", h.createDavSession)
	auth.Post("/settings/webdav/{sessionID}/delete", h.deleteDavSession)

	auth.Get("/settings/users", h.getUsers)
	auth.Post("/settings/users", h.createUser)
	auth.Post("/settings/users/{userID}/delete", h.deleteUser)
}

func (h *settingsHandler) String() string {
	return "web.settings"
}

func (h *settingsHandler) getBrowsersSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, session := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	webSessions, err := h.webSessions.GetAllForUser(ctx, session.UserID(), nil)
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the websessions: %w", err))
		return
	}

	h.response.WriteHTML(w, r, http.StatusOK, "settings/browsers.tmpl", map[string]interface{}{
		"isAdmin":        user.IsAdmin(),
		"currentSession": session,
		"webSessions":    webSessions,
	})
}

func (h *settingsHandler) getDavSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	davSessions, err := h.davSessions.GetAllForUser(ctx, user.ID(), &storage.PaginateCmd{Limit: 10})
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the davsessions: %w", err))
		return
	}

	folders, err := h.folders.GetAllUserFolders(ctx, user.ID(), nil)
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the folders: %w", err))
		return
	}

	h.response.WriteHTML(w, r, http.StatusOK, "settings/webdav.tmpl", map[string]interface{}{
		"isAdmin":     user.IsAdmin(),
		"davSessions": davSessions,
		"folders":     folders,
	})
}

func (h *settingsHandler) createDavSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	folders := []uuid.UUID{}
	for _, rawUUID := range strings.Split(r.FormValue("folders"), ",") {
		id, err := h.uuid.Parse(rawUUID)
		if err != nil {
			fmt.Fprintf(w, `<div class="alert alert-danger role="alert">invalid id: %s</div>`, err)
			return
		}

		folders = append(folders, id)
	}

	newSession, secret, err := h.davSessions.Create(ctx, &davsessions.CreateCmd{
		UserID:  user.ID(),
		Name:    r.FormValue("name"),
		Folders: folders,
	})
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	davSessions, err := h.davSessions.GetAllForUser(ctx, user.ID(), &storage.PaginateCmd{Limit: 10})
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the davsessions: %w", err))
		return
	}

	h.response.WriteHTML(w, r, http.StatusOK, "settings/webdav.tmpl", map[string]interface{}{
		"isAdmin":     user.IsAdmin(),
		"davSessions": davSessions,
		"newSession":  newSession,
		"secret":      secret,
	})
}

func (h *settingsHandler) deleteWebSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	err := h.webSessions.Delete(ctx, &websessions.DeleteCmd{
		UserID: user.ID(),
		Token:  chi.URLParam(r, "sessionToken"),
	})
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	w.Header().Set("Location", "/settings/browsers")
	w.WriteHeader(http.StatusFound)
}

func (h *settingsHandler) deleteDavSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	sessionID, err := h.uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		w.Write([]byte(`<div class="alert alert-danger role="alert">Invalid session id</div>`))
		return
	}

	err = h.davSessions.Delete(ctx, &davsessions.DeleteCmd{
		UserID:    user.ID(),
		SessionID: sessionID,
	})
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	w.Header().Set("Location", "/settings/webdav")
	w.WriteHeader(http.StatusFound)
}

func (h *settingsHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	if !user.IsAdmin() {
		w.Write([]byte(`<div class="alert alert-danger role="alert">Action reserved to admins</div>`))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	users, err := h.users.GetAll(ctx, &storage.PaginateCmd{
		StartAfter: map[string]string{"username": ""},
		Limit:      10,
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.response.WriteHTML(w, r, http.StatusOK, "settings/users.tmpl", map[string]interface{}{
		"isAdmin": user.IsAdmin(),
		"current": user,
		"users":   users,
	})
}

func (h *settingsHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	if !user.IsAdmin() {
		w.Write([]byte(`<div class="alert alert-danger role="alert">Action reserved to admins</div>`))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userToDelete, err := h.uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		w.Write([]byte(`<div class="alert alert-danger role="alert">Invalid id</div>`))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.users.AddToDeletion(ctx, userToDelete)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", "/settings/users")
	w.WriteHeader(http.StatusFound)
}

func (h *settingsHandler) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	user, err := h.users.Create(ctx, &users.CreateCmd{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		IsAdmin:  r.FormValue("role") == "admin",
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", "/settings/users")
	w.WriteHeader(http.StatusFound)
}

func (h *settingsHandler) getUserAndSession(w http.ResponseWriter, r *http.Request) (*users.User, *websessions.Session) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return nil, nil
	}

	user, err := h.users.GetByID(ctx, currentSession.UserID())
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return nil, nil
	}

	if user == nil {
		_ = h.webSessions.Logout(r, w)
		return nil, nil
	}

	return user, currentSession
}
