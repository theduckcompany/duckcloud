package oauth2

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/myminicloud/myminicloud/src/service/oauthclients"
	"github.com/myminicloud/myminicloud/src/service/oauthcodes"
	"github.com/myminicloud/myminicloud/src/service/oauthsessions"
	"github.com/myminicloud/myminicloud/src/tools"
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
