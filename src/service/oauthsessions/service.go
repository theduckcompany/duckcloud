package oauthsessions

import (
	"context"
	"fmt"
	"time"

	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/clock"
	"github.com/myminicloud/myminicloud/src/tools/errs"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

type (
	//go:generate mockery --name Storage
	Storage interface {
		Save(ctx context.Context, session *Session) error
		RemoveByAccessToken(ctx context.Context, access string) error
		RemoveByRefreshToken(ctx context.Context, refresh string) error
		GetByAccessToken(ctx context.Context, access string) (*Session, error)
		GetByRefreshToken(ctx context.Context, refresh string) (*Session, error)
	}

	// OauthSessionsService handling all the logic.
	OauthSessionsService struct {
		storage Storage
		clock   clock.Clock
	}
)

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
