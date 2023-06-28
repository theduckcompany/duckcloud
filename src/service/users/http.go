package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/go-chi/chi/v5"
)

// HTTPHandler handle all the HTTP request for the users
type HTTPHandler struct {
	service  Service
	response response.Writer
	jwt      jwt.Parser
}

func NewHTTPHandler(tools tools.Default, service Service) *HTTPHandler {
	return &HTTPHandler{
		service:  service,
		response: tools.ResWriter(),
		jwt:      tools.JWT(),
	}
}

// Register the http endpoints into the given mux server.
func (t *HTTPHandler) Register(r *chi.Mux) {
	r.Post("/users", t.createUser)
	r.Get("/users/me", t.getMyUser)
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
		t.response.WriteError(err, w, r)
		return
	}

	user, err := t.service.Create(r.Context(), &CreateUserRequest{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		t.response.WriteError(err, w, r)
		return
	}

	t.response.Write(w, r, &user, http.StatusCreated)
}

func (t *HTTPHandler) getMyUser(w http.ResponseWriter, r *http.Request) {
	type response struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"createdAt"`
	}

	token, err := t.jwt.FetchAccessToken(r)
	if err != nil {
		t.response.WriteError(err, w, r)
		return
	}

	user, err := t.service.GetByID(r.Context(), token.UserID)
	if err != nil {
		t.response.WriteError(err, w, r)
		return
	}

	t.response.Write(w, r, &response{
		ID:        string(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, http.StatusOK)
}
