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

	auth.Get("/settings", h.getBrowsersSessions(true))

	auth.Get("/settings/browsers", h.getBrowsersSessions(false))
	auth.Delete("/settings/browsers/{sessionToken}", h.deleteWebSession)

	auth.Get("/settings/webdav", h.getDavSessions)
	auth.Post("/settings/webdav", h.createDavSession)
	auth.Delete("/settings/webdav/{sessionID}", h.deleteDavSession)
}

func (h *settingsHandler) String() string {
	return "web.settings"
}

func (h *settingsHandler) getBrowsersSessions(withLayout bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		currentSession, err := h.webSessions.GetFromReq(r)
		if err != nil || currentSession == nil {
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusFound)
			return
		}

		webSessions, err := h.webSessions.GetUserSessions(ctx, currentSession.UserID())
		if err != nil {
			h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the websessions: %w", err))
			return
		}

		h.response.WriteHTML(w, http.StatusOK, "settings/browsers.tmpl", withLayout, map[string]interface{}{
			"currentSession": currentSession,
			"webSessions":    webSessions,
		})
	}
}

func (h *settingsHandler) getDavSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	davSessions, err := h.davSessions.GetAllForUser(ctx, currentSession.UserID(), &storage.PaginateCmd{Limit: 10})
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to fetch the davsessions: %w", err))
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "settings/webdav.tmpl", false, map[string]interface{}{
		"davSessions":   davSessions,
		"oauthSessions": []string{},
	})
}

func (h *settingsHandler) createDavSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	user, err := h.users.GetByID(ctx, currentSession.UserID())
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	session, secret, err := h.davSessions.Create(ctx, &davsessions.CreateCmd{
		UserID: currentSession.UserID(),
		Name:   r.FormValue("name"),
		FSRoot: user.RootFS(),
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "settings/show-dav-credentials.tmpl", false, map[string]interface{}{
		"session": session,
		"secret":  secret,
	})
}

func (h *settingsHandler) deleteWebSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	err = h.webSessions.Revoke(ctx, &websessions.RevokeCmd{
		UserID: currentSession.UserID(),
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

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	sessionID, err := h.uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	err = h.davSessions.Revoke(ctx, &davsessions.RevokeCmd{
		UserID:    currentSession.UserID(),
		SessionID: sessionID,
	})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`<div class="alert alert-danger role="alert">%s</div>`, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}
