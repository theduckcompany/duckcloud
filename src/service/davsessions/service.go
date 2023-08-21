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
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *DavSession) error
	GetByUsernamePassword(ctx context.Context, username, password string) (*DavSession, error)
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

func (s *DavSessionsService) Create(ctx context.Context, cmd *CreateCmd) (*DavSession, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	user, err := s.users.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to usersGetByID: %w", err)
	}

	if user == nil {
		return nil, errs.ValidationError(errors.New("userID: not found"))
	}

	rootInode, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     cmd.FSRoot,
		UserID:   user.ID(),
		FullName: "/",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inodes.Get: %w", err)
	}

	if rootInode == nil {
		return nil, errs.ValidationError(errors.New("rootFS: not found"))
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
		return nil, fmt.Errorf("failed to save the session: %w", err)
	}

	return &session, nil
}
