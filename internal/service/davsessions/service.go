package davsessions

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserIDNotMatching  = errors.New("user ids are not matching")
	ErrInvalidSpaceID     = errors.New("invalid spaceID")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *DavSession) error
	GetByUsernameAndPassword(ctx context.Context, username string, password secret.Text) (*DavSession, error)
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]DavSession, error)
	GetByID(ctx context.Context, sessionID uuid.UUID) (*DavSession, error)
	RemoveByID(ctx context.Context, sessionID uuid.UUID) error
}

type service struct {
	storage Storage
	spaces  spaces.Service
	uuid    uuid.Service
	clock   clock.Clock
}

func newService(storage Storage,
	spaces spaces.Service,
	tools tools.Tools,
) *service {
	return &service{storage, spaces, tools.UUID(), tools.Clock()}
}

func (s *service) Create(ctx context.Context, cmd *CreateCmd) (*DavSession, string, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, "", errs.Validation(err)
	}

	space, err := s.spaces.GetUserSpace(ctx, cmd.UserID, cmd.SpaceID)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, "", errs.Internal(fmt.Errorf("failed to get the space %q by id: %w", cmd.SpaceID, err))
	}

	if space == nil || !slices.Contains(space.Owners(), cmd.UserID) {
		return nil, "", errs.BadRequest(ErrInvalidSpaceID, "invalid spaces")
	}

	password := string(s.uuid.New())

	session := DavSession{
		id:        s.uuid.New(),
		userID:    cmd.UserID,
		name:      cmd.Name,
		username:  cmd.Username,
		password:  secret.NewText(hex.EncodeToString([]byte(password))),
		spaceID:   space.ID(),
		createdAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &session)
	if err != nil {
		return nil, "", errs.Internal(fmt.Errorf("failed to save the session: %w", err))
	}

	return &session, password, nil
}

func (s *service) Authenticate(ctx context.Context, username string, password secret.Text) (*DavSession, error) {
	res, err := s.storage.GetByUsernameAndPassword(ctx, username, secret.NewText(hex.EncodeToString([]byte(password.Raw()))))
	if errors.Is(err, errNotFound) {
		return nil, errs.BadRequest(ErrInvalidCredentials, "invalid credentials")
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByUsernameandPassword: %w", err))
	}

	return res, nil
}

func (s *service) GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *storage.PaginateCmd) ([]DavSession, error) {
	res, err := s.storage.GetAllForUser(ctx, userID, paginateCmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *service) Delete(ctx context.Context, cmd *DeleteCmd) error {
	err := cmd.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	session, err := s.storage.GetByID(ctx, cmd.SessionID)
	if errors.Is(err, errNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to GetByToken: %w", err)
	}

	if session.UserID() != cmd.UserID {
		return errs.NotFound(ErrUserIDNotMatching, "not found")
	}

	err = s.storage.RemoveByID(ctx, session.ID())
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to RemoveByID: %w", err))
	}

	return nil
}

func (s *service) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	davSessions, err := s.GetAllForUser(ctx, userID, nil)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetAllForUser: %w", err))
	}

	for _, session := range davSessions {
		err = s.Delete(ctx, &DeleteCmd{
			UserID:    userID,
			SessionID: session.ID(),
		})
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to Delete dav session %q: %w", session.ID(), err))
		}
	}

	return nil
}
