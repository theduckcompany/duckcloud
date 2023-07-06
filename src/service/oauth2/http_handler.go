package oauth2

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-oauth2/oauth2/v4"
	oerrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"golang.org/x/exp/slog"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

const (
	WebAppBaseURL = "http://localhost:8080"
)

// HTTPHandler handle all the oauth2 requests.
type HTTPHandler struct {
	logger   *slog.Logger
	jwt      jwt.Parser
	response response.Writer
	uuid     uuid.Service

	srv     *server.Server
	manager *manage.Manager

	users        users.Service
	clients      oauthclients.Service
	code         oauthcodes.Service
	oauthSession oauthsessions.Service
	webSession   *scs.SessionManager
}

// NewHTTPHandler setup a new Oauth2Server.
func NewHTTPHandler(
	db *sql.DB,
	tools tools.Tools,
	users users.Service,
	clients oauthclients.Service,
	code oauthcodes.Service,
	oauthSession oauthsessions.Service,
) *HTTPHandler {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MapTokenStorage(&tokenStorage{tools.UUID(), code, oauthSession})
	manager.MapClientStorage(&clientStorage{client: clients})
	manager.MapAccessGenerate(tools.JWT().GenerateAccess())

	webSession := scs.New()
	webSession.Store = newWebSessionStorage(db)

	srv := server.NewServer(&server.Config{
		TokenType:            "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code, oauth2.Token},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengePlain,
			oauth2.CodeChallengeS256,
		},
	}, manager)

	res := &HTTPHandler{
		logger:   tools.Logger(),
		response: tools.ResWriter(),
		jwt:      tools.JWT(),
		uuid:     tools.UUID(),

		srv:     srv,
		manager: manager,

		users:        users,
		clients:      clients,
		code:         code,
		oauthSession: oauthSession,
		webSession:   webSession,
	}

	srv.SetInternalErrorHandler(res.errorHandler)
	srv.SetResponseErrorHandler(res.responseErrorHandler)
	srv.SetUserAuthorizationHandler(res.userAuthorizationHandler)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	return res
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	// NOTE: There is the LoadAndSave session middleware here.
	auth := r.With(mids.StripSlashed, mids.Logger, h.webSession.LoadAndSave)

	// Web pages
	auth.HandleFunc("/login", h.handleLoginPage)
	auth.HandleFunc("/login", h.handleLoginPage)
	auth.HandleFunc("/forgot", h.handleForgotPage)
	auth.HandleFunc("/consent", h.handleConsentPage)

	// Actions
	auth.Post("/auth/logout", h.handleLogoutEndpoint)
	auth.HandleFunc("/auth/authorize", h.handleAuthorizationEndpoint)
	auth.HandleFunc("/auth/token", h.handleTokenEndpoint)
}

func (h *HTTPHandler) String() string {
	return "auth"
}

func (h *HTTPHandler) userAuthorizationHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil {
		return "", oerrors.ErrInvalidClient
	}

	userID := h.webSession.GetString(r.Context(), "userID")
	if userID == "" {
		// There is no available session so it's the first call to the authorization handler.
		//
		// Create a new session with a session_id, save all the form arguments in it
		// and redirect the user to the login with the session_id as argument
		if r.Form == nil {
			r.ParseForm()
		}

		w.Header().Set("Location", "/login?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)

		return "", nil
	}

	// The user already have a session with a userID. This means that it have been
	// authenticated in the past and we saved it.
	clientIDConsent := h.webSession.GetString(r.Context(), r.FormValue("consent"))
	if client.SkipValidation || clientIDConsent == r.FormValue("client_id") {
		// We can skip the validation so we directly authorize the user
		return userID, nil
	}

	w.Header().Set("Location", "/consent?"+r.Form.Encode())
	w.WriteHeader(http.StatusFound)
	return "", nil
}

func (h *HTTPHandler) handleLogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	token, err := h.jwt.FetchAccessToken(r)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	err = h.manager.RemoveAccessToken(r.Context(), token.Raw)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	h.webSession.Destroy(r.Context())

	h.response.WriteJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) handleTokenEndpoint(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleTokenRequest(w, r)
	if err != nil {
		h.logger.Error("OAUTH2 token failure: %s", err)
	}
}

func (h *HTTPHandler) handleAuthorizationEndpoint(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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

	inputs := map[string]string{}
	loginErrors := map[string]string{}

	inputs["username"] = r.FormValue("username")

	user, err := h.users.Authenticate(r.Context(), r.FormValue("username"), r.FormValue("password"))
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

	h.webSession.Put(r.Context(), "userID", string(user.ID))

	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil || client == nil {
		h.response.WriteJSON(w, http.StatusBadRequest, oerrors.ErrInvalidClient)
		return
	}

	if client.SkipValidation {
		w.Header().Set("Location", "/auth/authorize")
		w.WriteHeader(http.StatusFound)
		return
	}

	w.Header().Set("Location", "/consent")
	w.WriteHeader(http.StatusFound)
}

func (h *HTTPHandler) handleConsentPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	rawUserID := h.webSession.GetString(r.Context(), "userID")
	if rawUserID == "" {
		h.response.WriteJSON(w, http.StatusBadRequest, oerrors.ErrInvalidRequest)
		return
	}

	userID, _ := h.uuid.Parse(rawUserID)

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		h.response.WriteJSON(w, http.StatusBadRequest, fmt.Errorf("failed to find the user %q: %w", userID, err))
		return
	}

	clientID := r.FormValue("client_id")
	client, err := h.clients.GetByID(r.Context(), clientID)
	if err != nil || client == nil {
		h.response.WriteJSON(w, http.StatusBadRequest, oerrors.ErrInvalidRequest)
		return
	}

	consentToken := string(h.uuid.New())
	h.webSession.Put(r.Context(), consentToken, client.ID)
	r.Form.Add("consent", consentToken)

	h.response.WriteHTML(w, http.StatusOK, "auth/consent.html", map[string]interface{}{
		"clientName": client.Name,
		"username":   user.Username,
		"scope":      strings.Split(r.FormValue("scope"), ","),
		"redirect":   template.URL("/auth/authorize?" + r.Form.Encode()),
	})
}

func (h *HTTPHandler) responseErrorHandler(res *oerrors.Response) {
	h.logger.Error("OAUTH2 response error", slog.String("error", res.Error.Error()), slog.String("description", res.Description))
}

func (h *HTTPHandler) errorHandler(err error) *oerrors.Response {
	h.logger.Error("OAUTH2 internal Error", slog.String("error", err.Error()))
	return oerrors.NewResponse(fmt.Errorf("internal error"), http.StatusInternalServerError)
}

type debugLogger struct {
	logger *slog.Logger
}

func (l *debugLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(format, v...)
}
