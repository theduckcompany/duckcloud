package davsessions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserIDNotMatching  = errors.New("user ids are not matching")
	ErrInvalidFolderID    = errors.New("invalid folderID")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *DavSession) error
	GetByUsernameAndPassHash(ctx context.Context, username, password string) (*DavSession, error)
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]DavSession, error)
	GetByID(ctx context.Context, sessionID uuid.UUID) (*DavSession, error)
	RemoveByID(ctx context.Context, sessionID uuid.UUID) error
}

type DavSessionsService struct {
	storage Storage
	folders folders.Service
	uuid    uuid.Service
	clock   clock.Clock
}

func NewService(storage Storage,
	folders folders.Service,
	tools tools.Tools,
) *DavSessionsService {
	return &DavSessionsService{storage, folders, tools.UUID(), tools.Clock()}
}

func (s *DavSessionsService) Create(ctx context.Context, cmd *CreateCmd) (*DavSession, string, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, "", errs.Validation(err)
	}

	for _, folderID := range cmd.Folders {
		folder, err := s.folders.GetUserFolder(ctx, cmd.UserID, folderID)
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return nil, "", errs.Internal(fmt.Errorf("failed to get the folder %q by id: %w", folderID, err))
		}

		if folder == nil || !slices.Contains(folder.Owners(), cmd.UserID) {
			return nil, "", errs.BadRequest(ErrInvalidFolderID, "invalid folders")
		}
	}

	password := string(s.uuid.New())
	rawSha := sha256.Sum256([]byte(password))

	session := DavSession{
		id:        s.uuid.New(),
		userID:    cmd.UserID,
		name:      cmd.Name,
		username:  cmd.Username,
		password:  hex.EncodeToString(rawSha[:]),
		folders:   cmd.Folders,
		createdAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &session)
	if err != nil {
		return nil, "", errs.Internal(fmt.Errorf("failed to save the session: %w", err))
	}

	return &session, password, nil
}

func (s *DavSessionsService) Authenticate(ctx context.Context, username, password string) (*DavSession, error) {
	rawSha := sha256.Sum256([]byte(password))

	res, err := s.storage.GetByUsernameAndPassHash(ctx, username, hex.EncodeToString(rawSha[:]))
	if errors.Is(err, errNotFound) {
		return nil, errs.BadRequest(ErrInvalidCredentials, "invalid credentials")
	}
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByUsernameandPassHash: %w", err))
	}

	return res, nil
}

func (s *DavSessionsService) GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *storage.PaginateCmd) ([]DavSession, error) {
	res, err := s.storage.GetAllForUser(ctx, userID, paginateCmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *DavSessionsService) Delete(ctx context.Context, cmd *DeleteCmd) error {
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

func (s *DavSessionsService) DeleteAll(ctx context.Context, userID uuid.UUID) error {
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
