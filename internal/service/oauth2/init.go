package oauth2

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
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
) *service {
	return newService(tools, code, oauthSession, clients)
}
