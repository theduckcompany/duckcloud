package oauth2

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-session/session"
	"golang.org/x/exp/slog"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/service/users"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

const (
	WebAppBaseURL = "http://localhost:8080"
)

// HTTPHandler handle all the oauth2 requests.
type HTTPHandler struct {
	logger   *slog.Logger
	uuid     uuid.Service
	jwt      jwt.Parser
	users    users.Service
	client   oauthclients.Service
	response response.Writer
	srv      *server.Server
	session  oauthsessions.Service
}

// NewHTTPHandler setup a new Oauth2Server.
func NewHTTPHandler(
	tools tools.Tools,
	users users.Service,
	client oauthclients.Service,
	code oauthcodes.Service,
	session oauthsessions.Service) *HTTPHandler {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	uuid := tools.UUID()

	manager.MapTokenStorage(&tokenStorage{uuid, code, session})
	manager.MapClientStorage(&clientStorage{client: client})

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
	srv.SetClientInfoHandler(server.ClientFormHandler)

	res := &HTTPHandler{
		uuid:     uuid,
		logger:   tools.Logger(),
		response: tools.ResWriter(),
		jwt:      tools.JWT(),
		users:    users,
		client:   client,
		srv:      srv,
		session:  session,
	}

	srv.SetInternalErrorHandler(res.errorHandler)
	srv.SetResponseErrorHandler(res.responseErrorHandler)
	srv.SetUserAuthorizationHandler(res.userAuthorizationHandler)

	return res
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r *chi.Mux) {
	r.Get("/auth/login", h.printLoginPage)
	r.Post("/auth/login", h.handleLoginForm)
	r.Get("/auth/permissions", h.printPermissionsPage)
	r.Post("/auth/logout", h.handleLogoutEndpoint)
	r.HandleFunc("/auth/authorize", h.handleAuthorizationEndpoint)
	r.HandleFunc("/auth/token", h.handleTokenEndpoint)
}

func (h *HTTPHandler) String() string {
	return "auth"
}

func (h *HTTPHandler) userAuthorizationHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		return "", err
	}

	userID, ok := store.Get("LoggedInUserID")
	if !ok {
		// There is no session_id so it's the first call to the authorization handler.
		//
		// Create a new session with a session_id, save all the form arguments in it
		// and redirect the user to the login with the session_id as argument
		if r.Form == nil {
			r.ParseForm()
		}

		store.Set("ReturnUri", r.Form)
		store.Save()

		w.Header().Set("Location", "/auth/login")
		w.WriteHeader(http.StatusFound)
		return "", nil
	}

	return string(userID.(uuid.UUID)), nil
}

func (h *HTTPHandler) handleLogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	token, err := h.jwt.FetchAccessToken(r)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	err = h.session.RemoveByAccessToken(r.Context(), token.Raw)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	h.response.WriteJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) printPermissionsPage(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	session, ok := store.Get("ReturnUri")
	if !ok {
		// This is not session created yet. This append whan a user land directy on
		// the /auth/permissions without calling /auth/login first.
		w.Header().Set("Location", "/auth/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	form := session.(url.Values)

	userID, ok := store.Get("LoggedInUserID")
	if !ok {
		h.response.WriteJSON(w, http.StatusBadRequest, stderrors.New("user not authenticated"))
		return
	}

	user, err := h.users.GetByID(r.Context(), userID.(uuid.UUID))
	if err != nil || user == nil {
		h.response.WriteJSON(w, http.StatusBadRequest, fmt.Errorf("failed to find the user %q: %w", userID, err))
		return
	}

	h.response.WriteHTML(w, http.StatusOK, "auth/permissions.html", map[string]interface{}{
		"username": user.Username,
		"scope":    strings.Split(form.Get("scope"), ","),
	})
}

func (h *HTTPHandler) printLoginPage(w http.ResponseWriter, r *http.Request) {
	h.response.WriteHTML(w, http.StatusOK, "auth/login.html", nil)
}

func (h *HTTPHandler) handleLoginForm(w http.ResponseWriter, r *http.Request) {
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
	case stderrors.Is(err, users.ErrInvalidUsername):
		loginErrors["username"] = "User doesn't exists"
		status = http.StatusBadRequest
	case stderrors.Is(err, users.ErrInvalidPassword):
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

	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store.Set("LoggedInUserID", user.ID)
	store.Save()

	_, ok := store.Get("ReturnUri")
	if !ok {
		// There is not session created yet. This append when a user land directly on
		// /auth/login page without calling the /auth/authorize first (which would have
		// redirect him to the /login page).
		//
		// In that case we assume that the user want to connect to the web app and so we
		// will create the session with the correct client.
		client, err := h.client.GetByID(r.Context(), oauthclients.WebAppClientID)
		if err != nil {
			h.response.WriteJSONError(w, err)
			return
		}

		if client == nil {
			h.response.WriteJSONError(w, err)
			return
		}

		sessionID := string(h.uuid.New())
		form := url.Values{}
		form.Add("client_id", string(client.ID))
		form.Add("response_type", "code")
		// form.Add("redirect_uri", client.RedirectURI)
		form.Add("user_id", string(user.ID))
		form.Add("session_id", sessionID)
		form.Add("scope", client.Scopes.String())
		form.Add("state", string(h.uuid.New()))

		r.Form = form

		fmt.Printf("create a new web form: %+v\n", r.Form)
		store.Set("ReturnUri", r.Form)

		store.Save()
	}

	w.Header().Set("Location", "/auth/permissions")
	w.WriteHeader(http.StatusFound)
}

func (h *HTTPHandler) handleTokenEndpoint(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleTokenRequest(w, r)
	if err != nil {
		h.logger.Error("OAUTH2 token failure: %s", err)
	}
}

func (h *HTTPHandler) handleAuthorizationEndpoint(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
	}

	r.Form = form

	store.Delete("ReturnUri")
	store.Save()

	r.ParseForm()

	err = h.srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *HTTPHandler) responseErrorHandler(res *errors.Response) {
	h.logger.Error("OAUTH2 response error", slog.String("error", res.Error.Error()), slog.String("description", res.Description))
}

func (h *HTTPHandler) errorHandler(err error) *errors.Response {
	h.logger.Error("OAUTH2 internal Error", slog.String("error", err.Error()))
	return errors.NewResponse(fmt.Errorf("internal error"), http.StatusInternalServerError)
}

type debugLogger struct {
	logger *slog.Logger
}

func (l *debugLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(format, v...)
}
