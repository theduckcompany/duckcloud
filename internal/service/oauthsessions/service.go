package oauthsessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

//go:generate mockery --name storage
type storage interface {
	Save(ctx context.Context, session *Session) error
	RemoveByAccessToken(ctx context.Context, access secret.Text) error
	RemoveByRefreshToken(ctx context.Context, refresh secret.Text) error
	GetByAccessToken(ctx context.Context, access secret.Text) (*Session, error)
	GetByRefreshToken(ctx context.Context, refresh secret.Text) (*Session, error)
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Session, error)
}

// service handling all the logic.
type service struct {
	storage storage
	clock   clock.Clock
}

// newService create a new session service.
func newService(tools tools.Tools, storage storage) *service {
	return &service{storage, tools.Clock()}
}

func (s *service) Create(ctx context.Context, input *CreateCmd) (*Session, error) {
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

func (s *service) RemoveByAccessToken(ctx context.Context, access secret.Text) error {
	err := s.storage.RemoveByAccessToken(ctx, access)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *service) RemoveByRefreshToken(ctx context.Context, refresh secret.Text) error {
	err := s.storage.RemoveByRefreshToken(ctx, refresh)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *service) GetByAccessToken(ctx context.Context, access secret.Text) (*Session, error) {
	res, err := s.storage.GetByAccessToken(ctx, access)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *service) GetByRefreshToken(ctx context.Context, refresh secret.Text) (*Session, error) {
	res, err := s.storage.GetByRefreshToken(ctx, refresh)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *service) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Session, error) {
	res, err := s.storage.GetAllForUser(ctx, userID, cmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *service) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
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
