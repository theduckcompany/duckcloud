package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Peltoche/neurone/src/service/oauth2"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/router"
	"github.com/go-chi/chi/v5"
)

// HTTPHandler handle all the HTTP request for the users
type HTTPHandler struct {
	service   Service
	response  response.Writer
	oauth2Svc oauth2.Service
}

func NewHTTPHandler(tools tools.Tools, service Service, oauth2Svc oauth2.Service) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
		response:  tools.ResWriter(),
		oauth2Svc: oauth2Svc,
	}
}

// Register the http endpoints into the given mux server.
func (t *HTTPHandler) Register(r chi.Router, mids router.Middlewares) {
	users := r.With(mids.StripSlashed, mids.Logger, mids.CORS, mids.OnlyJSON)

	users.Post("/users", t.createUser)
	users.Get("/users/me", t.getMyUser)
}

func (h *HTTPHandler) String() string {
	return "users"
}

func (t *HTTPHandler) createUser(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input req

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		t.response.WriteJSONError(w, err)
		return
	}

	user, err := t.service.Create(r.Context(), &CreateCmd{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		t.response.WriteJSONError(w, err)
		return
	}

	t.response.WriteJSON(w, http.StatusCreated, &user)
}

func (t *HTTPHandler) getMyUser(w http.ResponseWriter, r *http.Request) {
	type response struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"createdAt"`
	}

	token, err := t.oauth2Svc.GetFromReq(r)
	if err != nil {
		t.response.WriteJSONError(w, err)
		return
	}

	user, err := t.service.GetByID(r.Context(), token.UserID)
	if err != nil {
		t.response.WriteJSONError(w, err)
		return
	}

	t.response.WriteJSON(w, http.StatusOK, &response{
		ID:        string(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}
