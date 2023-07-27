package oauth2

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-oauth2/oauth2/v4"
	oerrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"golang.org/x/exp/slog"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthconsents"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/service/websessions"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
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
	webSession   websessions.Service
	oauthConsent oauthconsents.Service
}

// NewHTTPHandler setup a new Oauth2Server.
func NewHTTPHandler(
	tools tools.Tools,
	users users.Service,
	clients oauthclients.Service,
	code oauthcodes.Service,
	oauthSession oauthsessions.Service,
	webSessions websessions.Service,
	oauthConsent oauthconsents.Service,
) *HTTPHandler {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MapTokenStorage(&tokenStorage{tools.UUID(), code, oauthSession})
	manager.MapClientStorage(&clientStorage{client: clients})
	manager.MapAccessGenerate(tools.JWT().GenerateAccess())

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
		webSession:   webSessions,
		oauthConsent: oauthConsent,
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
	auth := r.With(mids.RealIP, mids.StripSlashed, mids.Logger)

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
	_ = r.ParseForm()

	session, err := h.webSession.GetFromReq(r)
	if errors.Is(err, websessions.ErrMissingSessionToken) || session == nil {
		// There is no available session so it's the first call to the authorization handler.
		//
		// Create a new session with a session_id, save all the form arguments in it
		// and redirect the user to the login with the session_id as argument

		w.Header().Set("Location", "/login?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)

		return "", nil
	}
	if err != nil {
		return "", oerrors.ErrInvalidRequest
	}

	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil || client == nil {
		return "", oerrors.ErrInvalidClient
	}

	if client.SkipValidation {
		// We can skip the validation so we directly authorize the user
		return string(session.UserID), nil
	}

	err = h.oauthConsent.Check(r, client, session)
	if errors.Is(err, oauthconsents.ErrConsentNotFound) {
		w.Header().Set("Location", "/consent?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("invalid consent: %w", err)
	}

	return string(session.UserID), nil
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

	err = h.webSession.Logout(r, w)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

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
	_ = r.ParseForm()

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

	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil || client == nil {
		h.response.WriteJSONError(w, errs.BadRequest(ErrClientNotFound, "client not found"))
		return
	}

	session, err := h.webSession.Create(r.Context(), &websessions.CreateCmd{
		UserID:   string(user.ID),
		ClientID: client.ID,
		Req:      r,
	})
	if err != nil {
		h.response.WriteJSONError(w, fmt.Errorf("failed to create the websession: %w", err))
		return
	}

	// TODO: Handle the expiration time with the "Remember me" option
	c := http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, &c)

	if client.SkipValidation {
		w.Header().Set("Location", "/auth/authorize")
		w.WriteHeader(http.StatusFound)
		return
	}

	w.Header().Set("Location", "/consent?"+r.Form.Encode())
	w.WriteHeader(http.StatusFound)
}

func (h *HTTPHandler) handleConsentPage(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	session, err := h.webSession.GetFromReq(r)
	if errors.Is(err, websessions.ErrMissingSessionToken) || errors.Is(err, websessions.ErrSessionNotFound) {
		w.Header().Set("Location", "/login?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
		return
	}

	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	user, err := h.users.GetByID(r.Context(), session.UserID)
	if err != nil || user == nil {
		h.response.WriteJSONError(w, errs.BadRequest(fmt.Errorf("failed to find the user %q: %w", session.UserID, err), "user not found"))
		return
	}

	client, err := h.clients.GetByID(r.Context(), r.FormValue("client_id"))
	if err != nil {
		h.response.WriteJSONError(w, errs.BadRequest(err, "invalid request"))
		return
	}
	if client == nil {
		h.response.WriteJSONError(w, errs.BadRequest(ErrClientNotFound, "invalid request"))
		return
	}

	if r.Method == http.MethodPost {
		consent, err := h.oauthConsent.Create(r.Context(), &oauthconsents.CreateCmd{
			UserID:       user.ID,
			SessionToken: session.Token,
			ClientID:     client.ID,
			Scopes:       strings.Split(r.FormValue("scope"), ","),
		})
		if err != nil {
			h.response.WriteJSONError(w, fmt.Errorf("failed to create the consent: %w", err))
			return
		}

		r.Form.Add("consent_id", string(consent.ID))
		w.Header().Set("Location", "/auth/authorize?"+r.Form.Encode())
		w.WriteHeader(http.StatusFound)
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "auth/consent.html", map[string]interface{}{
		"clientName": client.Name,
		"username":   user.Username,
		"scope":      strings.Split(r.FormValue("scope"), ","),
		"redirect":   template.URL("/consent?" + r.Form.Encode()),
	})
}

func (h *HTTPHandler) responseErrorHandler(res *oerrors.Response) {
	h.logger.Error("OAUTH2 response error", slog.String("error", res.Error.Error()), slog.String("description", res.Description))
}

func (h *HTTPHandler) errorHandler(err error) *oerrors.Response {
	h.logger.Error("OAUTH2 internal Error", slog.String("error", err.Error()))
	return oerrors.NewResponse(fmt.Errorf("internal error"), http.StatusInternalServerError)
}
