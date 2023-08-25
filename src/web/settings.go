package web

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
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
	users       users.Service
	uuid        uuid.Service
}

func newSettingsHandler(
	tools tools.Tools,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	users users.Service,
) *settingsHandler {
	return &settingsHandler{
		response:    tools.ResWriter(),
		webSessions: webSessions,
		davSessions: davSessions,
		users:       users,
		uuid:        tools.UUID(),
	}
}

func (h *settingsHandler) Register(r chi.Router, mids router.Middlewares) {
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger)

	auth.Get("/settings", h.getBrowsersSessions)

	auth.Get("/settings/browsers", h.getBrowsersSessions)
	auth.Delete("/settings/browsers/{sessionToken}", h.deleteWebSession)

	auth.Get("/settings/webdav", h.getDavSessions)
	auth.Post("/settings/webdav", h.createDavSession)
	auth.Delete("/settings/webdav/{sessionID}", h.deleteDavSession)

	auth.Get("/settings/users", h.getUsers)
	auth.Delete("/settings/users/{userID}", h.deleteUser)
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

	webSessions, err := h.webSessions.GetUserSessions(ctx, session.UserID())
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the websessions: %w", err))
		return
	}

	fullPage := r.Header.Get("HX-Boosted") == ""

	h.response.WriteHTML(w, http.StatusOK, "settings/browsers.tmpl", fullPage, map[string]interface{}{
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

	fullPage := r.Header.Get("HX-Boosted") == ""

	h.response.WriteHTML(w, http.StatusOK, "settings/webdav.tmpl", fullPage, map[string]interface{}{
		"isAdmin":       user.IsAdmin(),
		"davSessions":   davSessions,
		"oauthSessions": []string{},
	})
}

func (h *settingsHandler) createDavSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	newSession, secret, err := h.davSessions.Create(ctx, &davsessions.CreateCmd{
		UserID: user.ID(),
		Name:   r.FormValue("name"),
		FSRoot: user.RootFS(),
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	fullPage := r.Header.Get("HX-Boosted") == ""

	h.response.WriteHTML(w, http.StatusOK, "settings/show-dav-credentials.tmpl", fullPage, map[string]interface{}{
		"session": newSession,
		"secret":  secret,
	})
}

func (h *settingsHandler) deleteWebSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	err := h.webSessions.Revoke(ctx, &websessions.RevokeCmd{
		UserID: user.ID(),
		Token:  chi.URLParam(r, "sessionToken"),
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *settingsHandler) deleteDavSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	sessionID, err := h.uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	err = h.davSessions.Revoke(ctx, &davsessions.RevokeCmd{
		UserID:    user.ID(),
		SessionID: sessionID,
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
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

	fullPage := r.Header.Get("HX-Boosted") == ""

	h.response.WriteHTML(w, http.StatusOK, "settings/users.tmpl", fullPage, map[string]interface{}{
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

	err = h.users.Delete(ctx, userToDelete)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
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
