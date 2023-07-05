package oauth2

import (
	"net/http"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

type Service interface {
	HandleOauthLogin(w http.ResponseWriter, r *http.Request, userID uuid.UUID)
}

func Init(tools tools.Tools, client oauthclients.Service) Service {
	return NewService(tools, client)
}
