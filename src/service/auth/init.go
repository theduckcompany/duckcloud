package auth

import (
	"net/http"

	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"golang.org/x/exp/slog"
)

type HTTPHandler struct {
	srv *server.Server
}

func NewHTTPHandler(log *logger.Logger) *HTTPHandler {
	manager := manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Error(err.Error(), slog.String("type", "internal"))
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Error(re.Error.Error(), slog.String("type", "response"))
	})

	return &HTTPHandler{srv}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/auth/authorize", h.authorizeHandler)
	mux.HandleFunc("/auth/token", h.tokenHandler)
	mux.HandleFunc("/auth/login", h.loginHandler)
}

func (h *HTTPHandler) String() string {
	return "auth"
}

func (h *HTTPHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		store.Set("LoggedInUserID", r.Form.Get("username"))
		store.Save()

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "static/login.html")
}

func (h *HTTPHandler) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *HTTPHandler) tokenHandler(w http.ResponseWriter, r *http.Request) {
	h.srv.HandleTokenRequest(w, r)
}
