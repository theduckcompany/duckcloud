package oauth2

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/myminicloud/myminicloud/src/service/oauthclients"
)

type clientStorage struct {
	client oauthclients.Service
}

func (t *clientStorage) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	return t.client.GetByID(ctx, id)
}
