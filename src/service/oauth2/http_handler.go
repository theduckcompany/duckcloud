package oauth2

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-oauth2/oauth2/v4"
	oerrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"

	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var ErrClientNotFound = errors.New("client not found")

// HTTPHandler handle all the oauth2 requests.
type HTTPHandler struct {
	logger   *slog.Logger
	response response.Writer
	uuid     uuid.Service

	srv *server.Server

	webSession   websessions.Service
	oauthConsent oauthconsents.Service
	clients      oauthclients.Service
}

// NewHTTPHandler setup a new Oauth2Server.
func NewHTTPHandler(
	tools tools.Tools,
	webSessions websessions.Service,
	oauthConsent oauthconsents.Service,
	clients oauthclients.Service,
	oaut2Svc Service,
) *HTTPHandler {
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
	}, oaut2Svc.manager())

	res := &HTTPHandler{
		logger:   tools.Logger(),
		response: tools.ResWriter(),
		uuid:     tools.UUID(),

		srv: srv,

		clients:      clients,
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
func (h *HTTPHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger, mids.CORS)
	}

	// Actions
	r.Post("/auth/logout", h.handleLogoutEndpoint)
	r.HandleFunc("/auth/authorize", h.handleAuthorizationEndpoint)
	r.HandleFunc("/auth/token", h.handleTokenEndpoint)
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

	clientID, err := h.uuid.Parse(r.FormValue("client_id"))
	if err != nil {
		return "", oerrors.ErrInvalidClient
	}

	client, err := h.clients.GetByID(r.Context(), clientID)
	if err != nil || client == nil {
		return "", oerrors.ErrInvalidClient
	}

	if client.SkipValidation() {
		// We can skip the validation so we directly authorize the user
		return string(session.UserID()), nil
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

	return string(session.UserID()), nil
}

func (h *HTTPHandler) handleLogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	err := h.webSession.Logout(r, w)
	if err != nil {
		h.response.WriteJSONError(w, r, err)
		return
	}

	h.response.WriteJSON(w, r, http.StatusOK, nil)
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

func (h *HTTPHandler) responseErrorHandler(res *oerrors.Response) {
	h.logger.Error("OAUTH2 response error", slog.String("error", res.Error.Error()), slog.String("description", res.Description))
}

func (h *HTTPHandler) errorHandler(err error) *oerrors.Response {
	h.logger.Error("OAUTH2 internal Error", slog.String("error", err.Error()))
	return oerrors.NewResponse(fmt.Errorf("internal error"), http.StatusInternalServerError)
}
