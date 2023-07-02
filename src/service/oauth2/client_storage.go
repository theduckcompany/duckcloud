package oauth2

import (
	"context"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"gopkg.in/oauth2.v3"
)

type clientStorage struct {
	client oauthclients.Service
}

func (t *clientStorage) GetByID(id string) (oauth2.ClientInfo, error) {
	return t.client.GetByID(context.Background(), id)
}
