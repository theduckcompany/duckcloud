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
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
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
func (t *UserService) Create(ctx context.Context, input *CreateCmd) (*User, error) {
	err := input.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	userWithEmail, err := t.storage.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if the email is already used: %w", err)
	}
	if userWithEmail != nil {
		return nil, errs.BadRequest(ErrAlreadyExists, "user already exists")
	}

	userWithSameUsername, err := t.storage.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check if the username is already used: %w", err)
	}
	if userWithSameUsername != nil {
		return nil, errs.BadRequest(ErrUsernameTaken, "username already taken")
	}

	newUserID := t.uuid.New()

	rootDir, err := t.inodes.BootstrapUser(ctx, newUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap the user inodes: %w", err)
	}

	hashedPassword, err := t.password.Encrypt(ctx, input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash the password: %w", err)
	}

	user := User{
		id:        newUserID,
		username:  input.Username,
		email:     input.Email,
		password:  hashedPassword,
		fsRoot:    rootDir.ID(),
		createdAt: t.clock.Now(),
	}

	err = t.storage.Save(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save the user: %w", err)
	}

	return &user, nil
}

// Authenticate return the user corresponding to the given username only if the password is correct.
func (t *UserService) Authenticate(ctx context.Context, username, userPassword string) (*User, error) {
	user, err := t.storage.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the user by its username: %w", err)
	}

	if user == nil {
		return nil, ErrInvalidUsername
	}

	err = t.password.Compare(ctx, user.password, userPassword)
	switch {
	case errors.Is(err, password.ErrMissmatchedPassword):
		return nil, ErrInvalidPassword
	case err != nil:
		return nil, fmt.Errorf("failed to compare the hash and the password: %w", err)
	default:
		return user, nil
	}
}

func (t *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return t.storage.GetByID(ctx, userID)
}
