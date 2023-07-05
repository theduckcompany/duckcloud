package oauth2

import (
	"net/http"
	"net/url"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/go-session/session"
)

type Oauth2Service struct {
	response response.Writer
	client   oauthclients.Service
	uuid     uuid.Service
}

func NewService(tools tools.Tools, client oauthclients.Service) *Oauth2Service {

	return &Oauth2Service{
		response: tools.ResWriter(),
		uuid:     tools.UUID(),
		client:   client,
	}
}

func (s *Oauth2Service) HandleOauthLogin(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store.Set("LoggedInUserID", userID)
	store.Save()

	_, ok := store.Get("ReturnUri")
	if !ok {
		// There is not session created yet. This append when a user land directly on
		// /login page without calling the /auth/authorize first (which would have
		// redirect him to the /login page).
		//
		// In that case we assume that the user want to connect to the web app and so we
		// will create the session with the correct client.
		client, err := s.client.GetByID(r.Context(), oauthclients.WebAppClientID)
		if err != nil {
			s.response.WriteJSONError(w, err)
			return
		}

		if client == nil {
			s.response.WriteJSONError(w, err)
			return
		}

		sessionID := string(s.uuid.New())
		form := url.Values{}
		form.Add("client_id", string(client.ID))
		form.Add("response_type", "code")
		form.Add("redirect_uri", client.RedirectURI)
		form.Add("user_id", string(userID))
		form.Add("session_id", sessionID)
		form.Add("scope", client.Scopes.String())
		form.Add("state", string(s.uuid.New()))

		r.Form = form

		store.Set("ReturnUri", r.Form)

		store.Save()
	}

	w.Header().Set("Location", "/consent")
	w.WriteHeader(http.StatusFound)
}
