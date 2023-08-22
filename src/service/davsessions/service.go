package davsessions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var ErrInvalidCredentials = fmt.Errorf("invalid credentials")

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *DavSession) error
	GetByUsernameAndPassHash(ctx context.Context, username, password string) (*DavSession, error)
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]DavSession, error)
}

type DavSessionsService struct {
	storage Storage
	inodes  inodes.Service
	users   users.Service
	uuid    uuid.Service
	clock   clock.Clock
}

func NewService(storage Storage, inodes inodes.Service, users users.Service, tools tools.Tools) *DavSessionsService {
	return &DavSessionsService{storage, inodes, users, tools.UUID(), tools.Clock()}
}

func (s *DavSessionsService) Create(ctx context.Context, cmd *CreateCmd) (*DavSession, string, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, "", errs.ValidationError(err)
	}

	user, err := s.users.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to usersGetByID: %w", err)
	}

	if user == nil {
		return nil, "", errs.ValidationError(errors.New("userID: not found"))
	}

	rootInode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     cmd.FSRoot,
		UserID:   user.ID(),
		FullName: "/",
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to inodes.Get: %w", err)
	}

	if rootInode == nil {
		return nil, "", errs.ValidationError(errors.New("rootFS: not found"))
	}

	password := string(s.uuid.New())
	rawSha := sha256.Sum256([]byte(password))

	session := DavSession{
		id:        s.uuid.New(),
		userID:    user.ID(),
		username:  user.Username(),
		password:  hex.EncodeToString(rawSha[:]),
		fsRoot:    rootInode.ID(),
		createdAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &session)
	if err != nil {
		return nil, "", fmt.Errorf("failed to save the session: %w", err)
	}

	return &session, password, nil
}

func (s *DavSessionsService) Authenticate(ctx context.Context, username, password string) (*DavSession, error) {
	rawSha := sha256.Sum256([]byte(password))

	res, err := s.storage.GetByUsernameAndPassHash(ctx, username, hex.EncodeToString(rawSha[:]))
	if err != nil {
		return nil, fmt.Errorf("failed to GetByUsernameandPassHash: %w", err)
	}

	if res == nil {
		return nil, errs.BadRequest(ErrInvalidCredentials, "invalid credentials")
	}

	return res, nil
}

func (s *DavSessionsService) GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *storage.PaginateCmd) ([]DavSession, error) {
	return s.storage.GetAllForUser(ctx, userID, paginateCmd)
}
