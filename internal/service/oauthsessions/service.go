package oauthsessions

import (
	"context"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
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

func (s *OauthSessionsService) Create(ctx context.Context, input *CreateCmd) (*Session, error) {
	err := input.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	now := s.clock.Now()

	session := Session{
		accessToken:      input.AccessToken,
		accessCreatedAt:  now,
		accessExpiresAt:  input.AccessExpiresAt,
		refreshToken:     input.RefreshToken,
		refreshCreatedAt: now,
		refreshExpiresAt: input.RefreshExpiresAt,
		clientID:         input.ClientID,
		userID:           input.UserID,
		scope:            input.Scope,
	}

	err = s.storage.Save(ctx, &session)
	if err != nil {
		return nil, fmt.Errorf("failed to save the refresh session: %w", err)
	}

	return &session, nil
}

func (s *OauthSessionsService) RemoveByAccessToken(ctx context.Context, access string) error {
	return s.storage.RemoveByAccessToken(ctx, access)
}

func (s *OauthSessionsService) RemoveByRefreshToken(ctx context.Context, refresh string) error {
	return s.storage.RemoveByRefreshToken(ctx, refresh)
}

func (s *OauthSessionsService) GetByAccessToken(ctx context.Context, access string) (*Session, error) {
	return s.storage.GetByAccessToken(ctx, access)
}

func (s *OauthSessionsService) GetByRefreshToken(ctx context.Context, refresh string) (*Session, error) {
	return s.storage.GetByRefreshToken(ctx, refresh)
}

func (s *OauthSessionsService) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error) {
	return s.storage.GetAllForUser(ctx, userID, cmd)
}

func (s *OauthSessionsService) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.GetAllForUser(ctx, userID, nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllForUser: %w", err)
	}

	for _, session := range sessions {
		err = s.RemoveByAccessToken(ctx, session.AccessToken())
		if err != nil {
			return fmt.Errorf("faile to RemoveByAccessToken: %w", err)
		}
	}

	return nil
}
