package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/password"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrAlreadyExists     = fmt.Errorf("user already exists")
	ErrUsernameTaken     = fmt.Errorf("username taken")
	ErrInvalidUsername   = fmt.Errorf("invalid username")
	ErrInvalidPassword   = fmt.Errorf("invalid password")
	ErrLastAdmin         = fmt.Errorf("can't remove the last admin")
	ErrInvalidStatus     = fmt.Errorf("invalid status")
	ErrUnauthorizedSpace = fmt.Errorf("unauthorized space")
)

// storage encapsulates the logic to access user from the data source.
//
//go:generate mockery --name storage
type storage interface {
	Save(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetAll(ctx context.Context, cmd *sqlstorage.PaginateCmd) ([]User, error)
	HardDelete(ctx context.Context, userID uuid.UUID) error
	Patch(ctx context.Context, userID uuid.UUID, fields map[string]any) error
}

// service handling all the logic.
type service struct {
	storage   storage
	clock     clock.Clock
	uuid      uuid.Service
	password  password.Password
	scheduler scheduler.Service
}

// newService create a new user service.
func newService(tools tools.Tools, storage storage, scheduler scheduler.Service) *service {
	return &service{
		storage,
		tools.Clock(),
		tools.UUID(),
		tools.Password(),
		scheduler,
	}
}

func (s *service) Bootstrap(ctx context.Context) (*User, error) {
	newUserID := s.uuid.New()
	return s.createUser(ctx, newUserID, BoostrapUsername, secret.NewText(BoostrapPassword), true, newUserID)
}

// Create will create and register a new user.
func (s *service) Create(ctx context.Context, cmd *CreateCmd) (*User, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	userWithSameUsername, err := s.storage.GetByUsername(ctx, cmd.Username)
	if err != nil && !errors.Is(err, errNotFound) {
		return nil, errs.Internal(fmt.Errorf("failed to GetByUsername: %w", err))
	}

	if userWithSameUsername != nil {
		return nil, errs.BadRequest(ErrUsernameTaken, "username already taken")
	}

	newUserID := s.uuid.New()
	return s.createUser(ctx, newUserID, cmd.Username, cmd.Password, cmd.IsAdmin, cmd.CreatedBy.id)
}

func (s *service) createUser(ctx context.Context, newUserID uuid.UUID, username string, password secret.Text, isAdmin bool, createdBy uuid.UUID) (*User, error) {
	hashedPassword, err := s.password.Encrypt(ctx, password)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to hash the password: %w", err))
	}

	now := s.clock.Now()

	user := User{
		id:                newUserID,
		username:          username,
		isAdmin:           isAdmin,
		password:          hashedPassword,
		status:            Initializing,
		passwordChangedAt: now,
		createdAt:         now,
		createdBy:         createdBy,
	}

	err = s.storage.Save(ctx, &user)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to save the user: %w", err))
	}

	ctx = context.WithoutCancel(ctx)
	// XXX:MULTI-WRITE
	//
	// This multi write can lead to a user blocked into the "pending" state. In order to reduce
	// those risks an rollback is done in case of error.
	//
	// TODO: Fix this with a commit systeme
	err = s.scheduler.RegisterUserCreateTask(ctx, &scheduler.UserCreateArgs{UserID: user.ID()})
	if err != nil {
		// Rollback the newly created user in order to avoid any invalid state.
		_ = s.storage.HardDelete(ctx, newUserID)
		return nil, fmt.Errorf("failed to RegisterUserCreateTask: %w", err)
	}

	return &user, nil
}

func (s *service) UpdateUserPassword(ctx context.Context, cmd *UpdatePasswordCmd) error {
	user, err := s.GetByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	hashedPassword, err := s.password.Encrypt(ctx, cmd.NewPassword)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to hash the password: %w", err))
	}

	err = s.storage.Patch(ctx, user.ID(), map[string]any{
		"password":            hashedPassword,
		"password_changed_at": sqlstorage.SQLTime(s.clock.Now()),
	})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to patch the user: %w", err))
	}

	return nil
}

func (s *service) MarkInitAsFinished(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	if user.status != Initializing {
		return nil, errs.Internal(ErrInvalidStatus)
	}

	user.status = Active

	err = s.storage.Patch(ctx, userID, map[string]any{"status": Active})
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to patch the user: %w", err))
	}

	return user, nil
}

func (s *service) GetAllWithStatus(ctx context.Context, status Status, cmd *sqlstorage.PaginateCmd) ([]User, error) {
	allUsers, err := s.GetAll(ctx, cmd)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetAll users: %w", err))
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
func (s *service) Authenticate(ctx context.Context, username string, userPassword secret.Text) (*User, error) {
	user, err := s.storage.GetByUsername(ctx, username)
	if errors.Is(err, errNotFound) {
		return nil, errs.BadRequest(ErrInvalidUsername)
	}
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetbyUsername: %w", err))
	}

	ok, err := s.password.Compare(ctx, user.password, userPassword)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed password compare: %w", err))
	}

	if !ok {
		return nil, errs.BadRequest(ErrInvalidPassword)
	}

	return user, nil
}

func (s *service) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	res, err := s.storage.GetByID(ctx, userID)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *service) GetAll(ctx context.Context, paginateCmd *sqlstorage.PaginateCmd) ([]User, error) {
	res, err := s.storage.GetAll(ctx, paginateCmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *service) AddToDeletion(ctx context.Context, userID uuid.UUID) error {
	user, err := s.GetByID(ctx, userID)
	if errors.Is(err, errNotFound) {
		return errs.NotFound(err)
	}

	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	if user.IsAdmin() {
		users, err := s.GetAll(ctx, nil)
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to GetAll: %w", err))
		}

		if isTheLastAdmin(users) {
			return errs.Unauthorized(ErrLastAdmin, "you are the last admin, you account can't be removed")
		}
	}

	// XXX:MULTI-WRITE
	err = s.scheduler.RegisterUserDeleteTask(ctx, &scheduler.UserDeleteArgs{
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to RegisterUserDeleteTask: %w", err)
	}

	ctx = context.WithoutCancel(ctx)

	// XXX:MULTI-WRITE
	//
	// This multi-write is not really dangerous. The "Deleting" stats patch allows to remove the
	// access user access immediately but as soon the task is executed every data related to this
	// user will be removed so there is no data corruptions.
	err = s.storage.Patch(ctx, userID, map[string]any{"status": Deleting})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to Patch the user: %w", err))
	}

	return nil
}

func (s *service) HardDelete(ctx context.Context, userID uuid.UUID) error {
	res, err := s.storage.GetByID(ctx, userID)
	if errors.Is(err, errNotFound) {
		return nil
	}
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetDeleted: %w", err))
	}

	if res.status != Deleting {
		return errs.Internal(ErrInvalidStatus)
	}

	err = s.storage.HardDelete(ctx, userID)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to HardDelete: %w", err))
	}

	return nil
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
