package users

import (
	"context"
	"fmt"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

var (
	ErrAlreadyExists       = fmt.Errorf("user already exists")
	ErrUsernameTaken       = fmt.Errorf("username taken")
	ErrInvalidUserPassword = fmt.Errorf("invalid pair user/password")
)

// Storage encapsulates the logic to access user from the data source.
type (
	Storage interface {
		Save(ctx context.Context, user *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByUsername(ctx context.Context, username string) (*User, error)
		GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	}

	// service handling all the logic.
	UserService struct {
		storage Storage
		clock   clock.Clock
		uuid    uuid.Service
	}
)

// NewService create a new user service.
func NewService(tools tools.Tools, storage Storage) *UserService {
	return &UserService{storage, tools.Clock, tools.UUID}
}

// Create will create and register a new user.
func (t *UserService) Create(ctx context.Context, input *CreateUserRequest) (*User, error) {
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

	user := User{
		ID:        t.uuid.New(),
		Username:  input.Username,
		Email:     input.Email,
		password:  input.Password,
		CreatedAt: t.clock.Now(),
	}

	err = t.storage.Save(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save the user: %w", err)
	}

	return &user, nil
}

// Authenticate return the user corresponding to the given username only if the password is correct.
func (t *UserService) Authenticate(ctx context.Context, username, password string) (*User, error) {
	user, err := t.storage.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the user by its username: %w", err)
	}

	if user == nil || user.password != password {
		return nil, ErrInvalidUserPassword
	}

	return user, nil
}

func (t *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return t.storage.GetByID(ctx, userID)
}
