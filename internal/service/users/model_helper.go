package users

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeUserBuilder struct {
	t    testing.TB
	user *User
}

func NewFakeUser(t testing.TB) *FakeUserBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()
	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeUserBuilder{
		t: t,
		user: &User{
			id:                uuidProvider.New(),
			createdAt:         createdAt,
			passwordChangedAt: createdAt,
			username:          gofakeit.Username(),
			password:          secret.NewText(gofakeit.Password(true, true, true, false, false, 8)),
			status:            Active,
			createdBy:         uuidProvider.New(),
			isAdmin:           false,
		},
	}
}

func (f *FakeUserBuilder) WithPassword(password string) *FakeUserBuilder {
	f.user.password = secret.NewText(password)

	return f
}

func (f *FakeUserBuilder) WithUsername(username string) *FakeUserBuilder {
	f.user.username = username

	return f
}

func (f *FakeUserBuilder) WithAdminRole() *FakeUserBuilder {
	f.user.isAdmin = true

	return f
}

func (f *FakeUserBuilder) WithStatus(status Status) *FakeUserBuilder {
	f.user.status = status

	return f
}

func (f *FakeUserBuilder) Build() *User {
	return f.user
}

func (f *FakeUserBuilder) BuildAndStore(ctx context.Context, db *sql.DB) *User {
	f.t.Helper()

	tools := tools.NewToolboxForTest(f.t)
	storage := newSqlStorage(db, tools)

	err := storage.Save(ctx, f.user)
	require.NoError(f.t, err)

	return f.user
}
