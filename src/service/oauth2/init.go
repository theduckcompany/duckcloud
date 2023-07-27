package oauth2

import (
	"net/http"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/go-oauth2/oauth2/v4/manage"
)

//go:generate mockery --name Service
type Service interface {
	GetFromReq(r *http.Request) (*Token, error)
	manager() *manage.Manager
}

func Init(
	tools tools.Tools,
	code oauthcodes.Service,
	oauthSession oauthsessions.Service,
	clients oauthclients.Service,
) *Oauth2Service {
	return NewService(tools, code, oauthSession, clients)
}
