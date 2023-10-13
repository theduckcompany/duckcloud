package oauth2

import (
	"context"
	"errors"

	"github.com/go-oauth2/oauth2/v4"
	oautherrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
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
	res, err := t.client.GetByID(ctx, uuid)
	if errors.Is(err, errs.ErrNotFound) {
		return nil, oautherrors.ErrInvalidClient
	}

	return res, nil
}
