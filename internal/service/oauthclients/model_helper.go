package oauthclients

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeClientBuilder struct {
	t      testing.TB
	client *Client
}

func NewFakeClient(t testing.TB) *FakeClientBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()
	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeClientBuilder{
		t: t,
		client: &Client{
			id:             uuidProvider.New(),
			name:           gofakeit.Name(),
			secret:         gofakeit.Password(true, true, true, false, false, 8),
			redirectURI:    gofakeit.URL(),
			userID:         uuidProvider.New(),
			createdAt:      createdAt,
			scopes:         Scopes{"scope-a", "scope-b"},
			public:         false,
			skipValidation: false,
		},
	}
}

func (f *FakeClientBuilder) SkipValidation() *FakeClientBuilder {
	f.client.skipValidation = true

	return f
}

func (f *FakeClientBuilder) CreatedBy(user *users.User) *FakeClientBuilder {
	f.client.userID = user.ID()

	return f
}

func (f *FakeClientBuilder) Build() *Client {
	return f.client
}

func (f *FakeClientBuilder) BuildAndStore(ctx context.Context, db *sql.DB) *Client {
	f.t.Helper()

	storage := newSqlStorage(db)

	err := storage.Save(ctx, f.client)
	require.NoError(f.t, err)

	return f.client
}
