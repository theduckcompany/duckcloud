package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Peltoche/neurone/src/service/oauth2"
	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/go-session/session"
)

type HTTPHandler struct {
	response     response.Writer
	users        users.Service
	oauth2       oauth2.Service
	oauthclients oauthclients.Service
}

func NewHTTPHandler(
	tools tools.Tools,
	users users.Service,
	oauth2 oauth2.Service,
	oauthclients oauthclients.Service,
) *HTTPHandler {
	return &HTTPHandler{
		response:     tools.ResWriter(),
		users:        users,
		oauth2:       oauth2,
		oauthclients: oauthclients,
	}
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	web := r.With(mids.StripSlashed)

	web.HandleFunc("/login", h.handleLoginPage)
	web.HandleFunc("/forgot", h.handleForgotPage)
	web.HandleFunc("/consent", h.handleConsentPage)
}

func (h *HTTPHandler) String() string {
	return "web"
}

func (h *HTTPHandler) handleForgotPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.response.WriteHTML(w, http.StatusOK, "auth/forgot.html", nil)
		return
	}

	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented yet!"))
}

func (h *HTTPHandler) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.response.WriteHTML(w, http.StatusOK, "auth/login.html", nil)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.response.WriteJSONError(w, errs.BadRequest(err, "invalid form url"))
		return
	}

	inputs := map[string]string{}
	loginErrors := map[string]string{}

	inputs["username"] = r.Form.Get("username")

	user, err := h.users.Authenticate(r.Context(), r.Form.Get("username"), r.Form.Get("password"))
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
		h.response.WriteHTML(w, status, "auth/login.html", map[string]interface{}{
			"inputs": inputs,
			"errors": loginErrors,
		})
		return
	}

	h.oauth2.HandleOauthLogin(w, r, user.ID)
}

func (h *HTTPHandler) handleConsentPage(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	session, ok := store.Get("ReturnUri")
	if !ok {
		// This is not session created yet. This append whan a user land directy on
		// the /consent without calling /login first.
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	form := session.(url.Values)

	userID, ok := store.Get("LoggedInUserID")
	if !ok {
		h.response.WriteJSON(w, http.StatusBadRequest, errors.New("user not authenticated"))
		return
	}

	user, err := h.users.GetByID(r.Context(), userID.(uuid.UUID))
	if err != nil || user == nil {
		h.response.WriteJSON(w, http.StatusBadRequest, fmt.Errorf("failed to find the user %q: %w", userID, err))
		return
	}

	clientID := form.Get("client_id")
	client, err := h.oauthclients.GetByID(r.Context(), clientID)
	if err != nil {
		h.response.WriteJSON(w, http.StatusBadRequest, fmt.Errorf("failed to find the client %q: %w", clientID, err))
		return
	}

	if client == nil {
		h.response.WriteJSON(w, http.StatusBadRequest, fmt.Errorf("client %q doesn't exists", clientID))
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "auth/consent.html", map[string]interface{}{
		"clientName": client.Name,
		"username":   user.Username,
		"scope":      strings.Split(form.Get("scope"), ","),
	})
}
