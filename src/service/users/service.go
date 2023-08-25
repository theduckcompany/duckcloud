package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/password"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	ErrAlreadyExists   = fmt.Errorf("user already exists")
	ErrUsernameTaken   = fmt.Errorf("username taken")
	ErrInvalidUsername = fmt.Errorf("invalid username")
	ErrInvalidPassword = fmt.Errorf("invalid password")
)

// Storage encapsulates the logic to access user from the data source.
//
//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	HardDelete(ctx context.Context, userID uuid.UUID) error
	GetDeletedUsers(ctx context.Context, limit int) ([]User, error)
}

// service handling all the logic.
type UserService struct {
	storage  Storage
	clock    clock.Clock
	uuid     uuid.Service
	password password.Password
	inodes   inodes.Service
}

// NewService create a new user service.
func NewService(tools tools.Tools, storage Storage, inodes inodes.Service) *UserService {
	return &UserService{storage, tools.Clock(), tools.UUID(), tools.Password(), inodes}
}

// Create will create and register a new user.
func (s *UserService) Create(ctx context.Context, cmd *CreateCmd) (*User, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	userWithSameUsername, err := s.storage.GetByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check if the username is already used: %w", err)
	}
	if userWithSameUsername != nil {
		return nil, errs.BadRequest(ErrUsernameTaken, "username already taken")
	}

	newUserID := s.uuid.New()

	rootDir, err := s.inodes.BootstrapUser(ctx, newUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap the user inodes: %w", err)
	}

	hashedPassword, err := s.password.Encrypt(ctx, cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash the password: %w", err)
	}

	user := User{
		id:        newUserID,
		username:  cmd.Username,
		isAdmin:   cmd.IsAdmin,
		password:  hashedPassword,
		fsRoot:    rootDir.ID(),
		createdAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save the user: %w", err)
	}

	return &user, nil
}

// Authenticate return the user corresponding to the given username only if the password is correct.
func (s *UserService) Authenticate(ctx context.Context, username, userPassword string) (*User, error) {
	user, err := s.storage.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the user by its username: %w", err)
	}

	if user == nil {
		return nil, ErrInvalidUsername
	}

	err = s.password.Compare(ctx, user.password, userPassword)
	switch {
	case errors.Is(err, password.ErrMissmatchedPassword):
		return nil, ErrInvalidPassword
	case err != nil:
		return nil, fmt.Errorf("failed to compare the hash and the password: %w", err)
	default:
		return user, nil
	}
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return s.storage.GetByID(ctx, userID)
}

func (s *UserService) GetAll(ctx context.Context, paginateCmd *storage.PaginateCmd) ([]User, error) {
	return s.storage.GetAll(ctx, paginateCmd)
}

func (s *UserService) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.storage.Delete(ctx, userID)
}

func (s *UserService) GetDeleted(ctx context.Context, limit int) ([]User, error) {
	return s.storage.GetDeletedUsers(ctx, limit)
}

func (s *UserService) HardDelete(ctx context.Context, userID uuid.UUID) error {
	return s.storage.HardDelete(ctx, userID)
}
