package oauth2

import (
	"context"
	"fmt"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"gopkg.in/oauth2.v3"
)

var ErrInvalidUUIDFormat = fmt.Errorf("invalid uuid format")

type clientStorage struct {
	uuid   uuid.Service
	client oauthclients.Service
}

func (t *clientStorage) GetByID(rawUUID string) (oauth2.ClientInfo, error) {
	id, err := t.uuid.Parse(rawUUID)
	if err != nil {
		return nil, ErrInvalidUUIDFormat
	}

	return t.client.GetByID(context.Background(), uuid.UUID(id))
}
