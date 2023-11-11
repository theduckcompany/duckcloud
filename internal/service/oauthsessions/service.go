package oauthsessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *Session) error
	RemoveByAccessToken(ctx context.Context, access secret.Text) error
	RemoveByRefreshToken(ctx context.Context, refresh secret.Text) error
	GetByAccessToken(ctx context.Context, access secret.Text) (*Session, error)
	GetByRefreshToken(ctx context.Context, refresh secret.Text) (*Session, error)
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
		return nil, errs.Internal(fmt.Errorf("failed to Save: %w", err))
	}

	return &session, nil
}

func (s *OauthSessionsService) RemoveByAccessToken(ctx context.Context, access secret.Text) error {
	err := s.storage.RemoveByAccessToken(ctx, access)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *OauthSessionsService) RemoveByRefreshToken(ctx context.Context, refresh secret.Text) error {
	err := s.storage.RemoveByRefreshToken(ctx, refresh)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *OauthSessionsService) GetByAccessToken(ctx context.Context, access secret.Text) (*Session, error) {
	res, err := s.storage.GetByAccessToken(ctx, access)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *OauthSessionsService) GetByRefreshToken(ctx context.Context, refresh secret.Text) (*Session, error) {
	res, err := s.storage.GetByRefreshToken(ctx, refresh)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *OauthSessionsService) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error) {
	res, err := s.storage.GetAllForUser(ctx, userID, cmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *OauthSessionsService) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.GetAllForUser(ctx, userID, nil)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetAllForUser: %w", err))
	}

	for _, session := range sessions {
		err = s.RemoveByAccessToken(ctx, session.AccessToken())
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to RemoveByAccessToken: %w", err))
		}
	}

	return nil
}
