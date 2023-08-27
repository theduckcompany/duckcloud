package oauthsessions

import (
	"context"
	"fmt"
	"time"

	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *Session) error
	RemoveByAccessToken(ctx context.Context, access string) error
	RemoveByRefreshToken(ctx context.Context, refresh string) error
	GetByAccessToken(ctx context.Context, access string) (*Session, error)
	GetByRefreshToken(ctx context.Context, refresh string) (*Session, error)
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error)
}

// OauthSessionsService handling all the logic.
type OauthSessionsService struct {
	storage Storage
	clock   clock.Clock
}

// NewService create a new session service.
func NewService(tools tools.Tools, storage Storage) *OauthSessionsService {
	return &OauthSessionsService{storage, tools.Clock()}
}

func (t *OauthSessionsService) Create(ctx context.Context, input *CreateCmd) error {
	err := input.Validate()
	if err != nil {
		return errs.ValidationError(err)
	}

	now := time.Now()

	err = t.storage.Save(ctx, &Session{
		accessToken:      input.AccessToken,
		accessCreatedAt:  now,
		accessExpiresAt:  input.AccessExpiresAt,
		refreshToken:     input.RefreshToken,
		refreshCreatedAt: now,
		refreshExpiresAt: input.RefreshExpiresAt,
		clientID:         input.ClientID,
		userID:           input.UserID,
		scope:            input.Scope,
	})
	if err != nil {
		return fmt.Errorf("failed to save the refresh session: %w", err)
	}

	return nil
}

func (t *OauthSessionsService) RemoveByAccessToken(ctx context.Context, access string) error {
	return t.storage.RemoveByAccessToken(ctx, access)
}

func (t *OauthSessionsService) RemoveByRefreshToken(ctx context.Context, refresh string) error {
	return t.storage.RemoveByRefreshToken(ctx, refresh)
}

func (t *OauthSessionsService) GetByAccessToken(ctx context.Context, access string) (*Session, error) {
	return t.storage.GetByAccessToken(ctx, access)
}

func (t *OauthSessionsService) GetByRefreshToken(ctx context.Context, refresh string) (*Session, error) {
	return t.storage.GetByRefreshToken(ctx, refresh)
}

func (t *OauthSessionsService) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error) {
	return t.storage.GetAllForUser(ctx, userID, cmd)
}
