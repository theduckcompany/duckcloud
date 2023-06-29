package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

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
	logger          *slog.Logger
	uuid            uuid.Service
	jwt             jwt.Parser
	users           users.Service
	client          oauthclients.Service
	response        response.Writer
	srv             *server.Server
	session         oauthsessions.Service
	sessionsStorage *sync.Map
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
	manager.MapClientStorage(&clientStorage{uuid: tools.UUID(), client: client})
	manager.MapAccessGenerate(tools.JWT().GenerateAccess())

	srv := server.NewServer(&server.Config{
		TokenType:            "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code, oauth2.Token},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
		},
	}, manager)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	res := &HTTPHandler{
		sessionsStorage: new(sync.Map),
		uuid:            uuid,
		logger:          tools.Logger(),
		response:        tools.ResWriter(),
		jwt:             tools.JWT(),
		users:           users,
		client:          client,
		srv:             srv,
		session:         session,
	}

	srv.SetInternalErrorHandler(res.errorHandler)
	srv.SetResponseErrorHandler(res.responseErrorHandler)
	srv.SetUserAuthorizationHandler(res.userAuthorizationHandler)

	return res
}

// Register the http endpoints into the given mux server.
func (h *HTTPHandler) Register(r *chi.Mux) {
	r.Get("/login", h.printLoginPage)
	r.Post("/login", h.handleLoginEndpoint)
	r.Post("/logout", h.handleLogoutEndpoint)
	r.HandleFunc("/oauth2/token", h.handleTokenEndpoint)
	r.HandleFunc("/oauth2/authorization", h.handleAuthorizationEndpoint)
}

func (h *HTTPHandler) String() string {
	return "oauth2"
}

func (h *HTTPHandler) userAuthorizationHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	if r.Form == nil {
		err := r.ParseForm()
		if err != nil {
			return "", errors.ErrInvalidRequest
		}
	}

	if userID := r.Form.Get("user_id"); userID != "" {
		// A session exists and the user have been setup after a successful connexion.
		return userID, nil
	}

	sessionID := r.Form.Get("session_id")
	if sessionID == "" {
		// There is no session_id so it's the first call to the authorization handler.
		//
		// Create a new session with a session_id, save all the form arguments in it
		// and redirect the user to the login with the session_id as argument
		sessionID = string(h.uuid.New())
		h.sessionsStorage.Store(sessionID, r.Form)
	}

	q := url.Values{}
	u, _ := url.Parse(WebAppBaseURL + "/login")
	q.Add("session_id", string(sessionID))
	u.RawQuery = q.Encode()

	w.Header().Set("Location", u.String())
	w.WriteHeader(http.StatusFound)
	return "", nil
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

func (h *HTTPHandler) printLoginPage(w http.ResponseWriter, r *http.Request) {
	h.response.WriteHTML(w, http.StatusOK, "login.html", nil)
}

func (h *HTTPHandler) handleLoginEndpoint(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type res struct {
		SessionID string `json:"sessionID"`
	}

	var input req

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	user, err := h.users.Authenticate(r.Context(), input.Username, input.Password)
	if err != nil {
		h.response.WriteJSONError(w, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		h.response.WriteJSONError(w, errs.BadRequest(err, "invalid form url"))
		return
	}

	sessionID := r.Form.Get("session_id")
	if sessionID == "" {
		// There is not session created yet. This append when a user land directly on
		// /login page without calling the /oauth2/redirect first (which would have
		// redirect him to the /login page).
		//
		// In that case we assume that the user want to connect to the web app and so we
		// will create the session with the correct client.

		client, err := h.client.GetByID(r.Context(), oauthclients.WebAppClientID)
		if err != nil {
			h.response.WriteJSONError(w, err)
			return
		}

		sessionID = string(h.uuid.New())
		form := url.Values{}
		form.Add("client_id", string(client.ID))
		form.Add("response_type", "code")
		form.Add("redirect_uri", client.RedirectURI)
		form.Add("user_id", string(user.ID))
		form.Add("session_id", sessionID)
		form.Add("scope", client.Scopes.String())

		h.sessionsStorage.Store(sessionID, form)
	} else {
		res, ok := h.sessionsStorage.Load(sessionID)
		if !ok {
			w.WriteHeader(http.StatusConflict)
			return
		}
		form := res.(url.Values)
		form.Add("user_id", string(user.ID))
		h.sessionsStorage.Store(sessionID, form)

	}

	h.response.WriteJSON(w, http.StatusOK, &res{sessionID})
}

func (h *HTTPHandler) handleTokenEndpoint(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleTokenRequest(w, r)
	if err != nil {
		h.logger.Error("OAUTH2 token failure: %s", err)
	}
}

func (h *HTTPHandler) handleAuthorizationEndpoint(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, errors.ErrInvalidRequest.Error(), http.StatusBadRequest)
		return
	}

	if sessionID := r.Form.Get("session_id"); sessionID != "" {
		// A session_id is found. This means that a session already exists. This session have
		// been created by a first call to the authorization handler which have redirected to the
		// login page or directy by the login handler if the connexion page have been directly called.
		// In that case the login handler have assume that the client is the web app have a automatically
		// created a session for it.
		form, ok := h.sessionsStorage.Load(sessionID)
		if !ok {
			http.Error(w, errors.ErrInvalidRequest.Error(), http.StatusBadRequest)
			return
		}
		r.Form = form.(url.Values)
	}

	err = h.srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *HTTPHandler) responseErrorHandler(res *errors.Response) {
	h.logger.Error("OAUTH2 response error: %s: %s", res.Error, res.Description)
}

func (h *HTTPHandler) errorHandler(err error) *errors.Response {
	h.logger.Error("Internal Error: %s", err)
	return errors.NewResponse(fmt.Errorf("internal error"), http.StatusInternalServerError)
}

type debugLogger struct {
	logger *slog.Logger
}

func (l *debugLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(format, v...)
}
