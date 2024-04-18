package oauthcodes

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeCodeBuilder struct {
	t    testing.TB
	code *Code
}

func NewFakeCode(t testing.TB) *FakeCodeBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()
	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeCodeBuilder{
		t: t,
		code: &Code{
			code:            secret.NewText(gofakeit.Password(true, true, true, false, false, 8)),
			createdAt:       createdAt,
			expiresAt:       createdAt.Add(time.Hour),
			clientID:        string(uuidProvider.New()),
			userID:          string(uuidProvider.New()),
			redirectURI:     gofakeit.URL(),
			scope:           "scope-1,scope-2",
			challenge:       secret.NewText(gofakeit.Password(true, true, true, false, false, 8)),
			challengeMethod: "S256",
		},
	}
}

func (f *FakeCodeBuilder) WithClient(client *oauthclients.Client) *FakeCodeBuilder {
	f.code.clientID = client.GetID()

	return f
}

func (f *FakeCodeBuilder) CreatedBy(user *users.User) *FakeCodeBuilder {
	f.code.userID = string(user.ID())

	return f
}

func (f *FakeCodeBuilder) Build() *Code {
	return f.code
}

// func (f *FakeCodeBuilder) BuildAndStore(ctx context.Context, db sqlstorage.Querier) *Code {
// 	f.t.Helper()
//
// 	tools := tools.NewToolboxForTest(f.t)
// 	storage := newSqlStorage(db, tools)
//
// 	err := storage.Save(ctx, f.code)
// 	require.NoError(f.t, err)
//
// 	return f.code
// }
