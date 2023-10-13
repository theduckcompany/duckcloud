package oauth2

import (
	"fmt"
	"net/http"
	"strings"

	oautherrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Oauth2Service struct {
	m *manage.Manager
}

func NewService(
	tools tools.Tools,
	code oauthcodes.Service,
	oauthSession oauthsessions.Service,
	clients oauthclients.Service,
) *Oauth2Service {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MapTokenStorage(&tokenStorage{tools.UUID(), code, oauthSession})
	manager.MapClientStorage(&clientStorage{client: clients})

	return &Oauth2Service{}
}

func (s *Oauth2Service) manager() *manage.Manager {
	return s.m
}

func (s *Oauth2Service) GetFromReq(r *http.Request) (*Token, error) {
	accessToken, ok := s.bearerAuth(r)
	if !ok {
		return nil, oautherrors.ErrInvalidAccessToken
	}

	token, err := s.manager().LoadAccessToken(r.Context(), accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to load the access token: %w", err)
	}

	if token == nil {
		return nil, oautherrors.ErrInvalidAccessToken
	}

	return &Token{
		UserID: uuid.UUID(token.GetUserID()),
	}, nil
}

func (s *Oauth2Service) bearerAuth(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "
	token := ""

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = r.FormValue("access_token")
	}

	return token, token != ""
}
