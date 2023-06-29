package oauthsessions

import (
	"context"
	"fmt"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/errs"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

type (
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

func (t *OauthSessionsService) CreateSession(ctx context.Context, input *CreateSessionRequest) error {
	err := input.Validate()
	if err != nil {
		return errs.ValidationError(err)
	}

	now := time.Now()

	err = t.storage.Save(ctx, &Session{
		AccessToken:      input.AccessToken,
		AccessCreatedAt:  now,
		AccessExpiresAt:  input.AccessExpiresAt,
		RefreshToken:     input.RefreshToken,
		RefreshCreatedAt: now,
		RefreshExpiresAt: input.RefreshExpiresAt,
		ClientID:         input.ClientID,
		UserID:           input.UserID,
		Scope:            input.Scope,
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