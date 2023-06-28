package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Service_Create_success(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)

	now := time.Now()
	ctx := context.Background()

	user := User{
		ID:        uuid.UUID("some-id"),
		Username:  "some-username",
		Email:     "some@email.com",
		password:  "some-password",
		CreatedAt: now,
	}

	storage.On("GetByEmail", ctx, "some@email.com").Return(nil, nil).Once()
	storage.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

	app.UUIDMock.On("New").Return(uuid.UUID("some-id")).Once()
	app.ClockMock.On("Now").Return(now)

	storage.On("Save", ctx, &user).Return(nil).Once()

	coloniesMock.On("BootstrapUser", ctx, user.ID).Return(&colonies.Colony{}, nil).Once()

	res, err := service.Create(ctx, &CreateUserRequest{
		Username: "some-username",
		Email:    "some@email.com",
		Password: "some-password",
	})
	assert.NoError(t, err)
	assert.Equal(t, &user, res)
}

func Test_Service_Create_with_email_already_exists(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	storage.On("GetByEmail", ctx, "some@email.com").Return(&User{}, nil).Once()

	res, err := service.Create(ctx, &CreateUserRequest{
		Username: "some-username",
		Email:    "some@email.com",
		Password: "some-password",
	})
	assert.ErrorIs(t, err, ErrAlreadyExists)
	assert.ErrorIs(t, err, response.ErrBadRequest)
	assert.Nil(t, res)
}

func Test_Service_Create_with_username_taken(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	storage.On("GetByEmail", ctx, "some@email.com").Return(nil, nil).Once()
	storage.On("GetByUsername", ctx, "some-username").Return(&User{}, nil).Once()

	res, err := service.Create(ctx, &CreateUserRequest{
		Username: "some-username",
		Email:    "some@email.com",
		Password: "some-password",
	})
	assert.ErrorIs(t, err, ErrUsernameTaken)
	assert.ErrorIs(t, err, response.ErrBadRequest)
	assert.Nil(t, res)
}

func Test_Service_Create_with_a_database_error(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	storage.On("GetByEmail", ctx, "some@email.com").Return(nil, fmt.Errorf("some-error")).Once()

	res, err := service.Create(ctx, &CreateUserRequest{
		Username: "some-username",
		Email:    "some@email.com",
		Password: "some-password",
	})
	assert.ErrorContains(t, err, "some-error")
	assert.Nil(t, res)
}

func Test_Service_Authenticate_success(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	user := User{
		ID:        uuid.UUID("some-id"),
		Username:  "some-username",
		Email:     "some-email",
		CreatedAt: time.Now(),
		password:  "some-password",
	}

	storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()

	res, err := service.Authenticate(ctx, "some-username", "some-password")
	assert.NoError(t, err)
	assert.Equal(t, &user, res)
}

func Test_Service_Authenticate_with_invalid_username(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	storage.On("GetByUsername", ctx, "some-username").Return(nil, nil).Once()

	res, err := service.Authenticate(ctx, "some-username", "some-password")
	assert.ErrorIs(t, err, ErrInvalidUserPassword)
	assert.Nil(t, res)
}

func Test_Service_Authenticate_with_invalid_password(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	user := User{
		ID:        uuid.UUID("some-id"),
		Username:  "some-username",
		Email:     "some-email",
		CreatedAt: time.Now(),
		password:  "some-password",
	}

	storage.On("GetByUsername", ctx, "some-username").Return(&user, nil).Once()

	// Invalid password here
	res, err := service.Authenticate(ctx, "some-username", "some-invalid-password")
	assert.ErrorIs(t, err, ErrInvalidUserPassword)
	assert.Nil(t, res)
}

func Test_Service_GetByID_success(t *testing.T) {
	app := app.NewMock(t)
	storage := NewMockUserStorage(t)
	coloniesMock := NewMockColonyBootstraper(t)
	service := NewService(app, storage, coloniesMock)
	ctx := context.Background()

	user := User{
		ID:        uuid.UUID("some-id"),
		Username:  "some-username",
		Email:     "some-email",
		CreatedAt: time.Now(),
		password:  "some-password",
	}

	storage.On("GetByID", ctx, user.ID).Return(&user, nil).Once()

	res, err := service.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, &user, res)
}
