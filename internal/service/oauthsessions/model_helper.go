package oauthsessions

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeSessionBuilder struct {
	t       testing.TB
	session *Session
}

func NewFakeSession(t testing.TB) *FakeSessionBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()
	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeSessionBuilder{
		t: t,
		session: &Session{
			accessToken:      secret.NewText(gofakeit.Password(true, true, true, false, false, 8)),
			accessCreatedAt:  createdAt,
			accessExpiresAt:  createdAt.Add(time.Hour),
			refreshToken:     secret.NewText(gofakeit.Password(true, true, true, false, false, 8)),
			refreshCreatedAt: createdAt,
			refreshExpiresAt: createdAt.Add(time.Hour),
			clientID:         gofakeit.Name(),
			userID:           uuidProvider.New(),
			scope:            "scope-a,scope-b",
		},
	}
}

func (f *FakeSessionBuilder) WithClient(client *oauthclients.Client) *FakeSessionBuilder {
	f.session.clientID = client.GetID()

	return f
}

func (f *FakeSessionBuilder) CreatedBy(user *users.User) *FakeSessionBuilder {
	f.session.userID = user.ID()

	return f
}

func (f *FakeSessionBuilder) Build() *Session {
	return f.session
}

// func (f *FakeSessionBuilder) BuildAndStore(ctx context.Context, db sqlstorage.Querier) *Session {
// 	f.t.Helper()
//
// 	storage := newSqlStorage(db)
//
// 	err := storage.Save(ctx, f.session)
// 	require.NoError(f.t, err)
//
// 	return f.session
// }
