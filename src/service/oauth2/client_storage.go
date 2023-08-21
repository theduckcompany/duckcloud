package oauth2

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type clientStorage struct {
	client oauthclients.Service
	uuid   uuid.Service
}

func (t *clientStorage) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	uuid, err := t.uuid.Parse(id)
	if err != nil {
		return nil, nil
	}
	return t.client.GetByID(ctx, uuid)
}
