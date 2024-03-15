package spaces

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeSpaceBuilder struct {
	t     *testing.T
	space *Space
}

func NewFakeSpace(t *testing.T) *FakeSpaceBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()

	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeSpaceBuilder{
		t: t,
		space: &Space{
			id:        uuidProvider.New(),
			name:      gofakeit.Animal(),
			owners:    Owners{},
			createdAt: createdAt,
			createdBy: uuidProvider.New(),
		},
	}
}

func (f *FakeSpaceBuilder) WithName(name string) *FakeSpaceBuilder {
	f.space.name = name

	return f
}

func (f *FakeSpaceBuilder) CreatedBy(user *users.User) *FakeSpaceBuilder {
	f.space.createdBy = user.ID()

	return f
}

func (f *FakeSpaceBuilder) CreatedAt(at time.Time) *FakeSpaceBuilder {
	f.space.createdAt = at

	return f
}

func (f *FakeSpaceBuilder) WithOwners(users ...users.User) *FakeSpaceBuilder {
	owners := make(Owners, len(users))

	for i, elem := range users {
		owners[i] = elem.ID()
	}

	f.space.owners = owners

	return f
}

func (f *FakeSpaceBuilder) Build() *Space {
	return f.space
}

func (f *FakeSpaceBuilder) BuildAndStore(ctx context.Context, db *sql.DB) *Space {
	f.t.Helper()

	tools := tools.NewToolboxForTest(f.t)
	storage := newSqlStorage(db, tools)

	err := storage.Save(ctx, f.space)
	require.NoError(f.t, err)

	return f.space
}
