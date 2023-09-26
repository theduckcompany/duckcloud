package users

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/password"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrAlreadyExists      = fmt.Errorf("user already exists")
	ErrUsernameTaken      = fmt.Errorf("username taken")
	ErrInvalidUsername    = fmt.Errorf("invalid username")
	ErrInvalidPassword    = fmt.Errorf("invalid password")
	ErrLastAdmin          = fmt.Errorf("can't remove the last admin")
	ErrInvalidStatus      = fmt.Errorf("invalid status")
	ErrUnauthorizedFolder = fmt.Errorf("unauthorized folder")
)

// Storage encapsulates the logic to access user from the data source.
//
//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]User, error)
	HardDelete(ctx context.Context, userID uuid.UUID) error
	Patch(ctx context.Context, userID uuid.UUID, fields map[string]any) error
}

// service handling all the logic.
type UserService struct {
	storage  Storage
	clock    clock.Clock
	uuid     uuid.Service
	password password.Password
	folders  folders.Service
}

// NewService create a new user service.
func NewService(tools tools.Tools,
	storage Storage,
	folders folders.Service,
) *UserService {
	return &UserService{
		storage,
		tools.Clock(),
		tools.UUID(),
		tools.Password(),
		folders,
	}
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

	hashedPassword, err := s.password.Encrypt(ctx, cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash the password: %w", err)
	}

	user := User{
		id:              newUserID,
		username:        cmd.Username,
		defaultFolderID: "",
		isAdmin:         cmd.IsAdmin,
		password:        hashedPassword,
		createdAt:       s.clock.Now(),
		status:          "initializing",
	}

	err = s.storage.Save(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save the user: %w", err)
	}

	return &user, nil
}

func (s *UserService) SetDefaultFolder(ctx context.Context, user User, folder *folders.Folder) (*User, error) {
	if !slices.Contains[[]uuid.UUID, uuid.UUID](folder.Owners(), user.ID()) {
		return nil, ErrUnauthorizedFolder
	}

	user.defaultFolderID = folder.ID()

	err := s.storage.Patch(ctx, user.ID(), map[string]any{"default_folder": folder.ID()})
	if err != nil {
		return nil, fmt.Errorf("failed to patch the user: %w", err)
	}

	return &user, nil
}

func (s *UserService) MarkInitAsFinished(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByID: %w", err)
	}

	if user.status != "initializing" {
		return nil, ErrInvalidStatus
	}

	user.status = "active"

	err = s.storage.Patch(ctx, userID, map[string]any{"status": "active"})
	if err != nil {
		return nil, fmt.Errorf("failed to patch the user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetAllWithStatus(ctx context.Context, status string, cmd *storage.PaginateCmd) ([]User, error) {
	allUsers, err := s.GetAll(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAll users: %w", err)
	}

	res := []User{}
	for _, user := range allUsers {
		if user.status == status {
			res = append(res, user)
		}
	}

	return res, nil
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

func (s *UserService) AddToDeletion(ctx context.Context, userID uuid.UUID) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	if user == nil {
		return nil
	}

	if user.IsAdmin() {
		users, err := s.GetAll(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to GetAll: %w", err)
		}

		if isTheLastAdmin(users) {
			return errs.Unauthorized(ErrLastAdmin, "you are the last admin, you account can't be removed")
		}
	}

	err = s.storage.Patch(ctx, userID, map[string]any{"status": "deleting"})
	if err != nil {
		return fmt.Errorf("failed to patch the user: %w", err)
	}

	return nil
}

func (s *UserService) HardDelete(ctx context.Context, userID uuid.UUID) error {
	res, err := s.storage.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to GetDeleted: %w", err)
	}

	if res == nil {
		return nil
	}

	if res.status != "deleting" {
		return ErrInvalidStatus
	}

	return s.storage.HardDelete(ctx, userID)
}

func isTheLastAdmin(users []User) bool {
	nbAdmin := 0

	for _, user := range users {
		if user.IsAdmin() {
			nbAdmin++
		}
	}

	return nbAdmin <= 1
}
