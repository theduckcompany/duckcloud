package oauthconsents

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeConsentBuilder struct {
	t       testing.TB
	consent *Consent
}

func NewFakeConsent(t testing.TB) *FakeConsentBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()
	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeConsentBuilder{
		t: t,
		consent: &Consent{
			createdAt:    createdAt,
			id:           uuidProvider.New(),
			userID:       uuidProvider.New(),
			sessionToken: gofakeit.Password(true, true, true, false, false, 8),
			clientID:     string(uuidProvider.New()),
			scopes:       []string{"scope-a", "scope-b"},
		},
	}
}

func (f *FakeConsentBuilder) WithClient(client *oauthclients.Client) *FakeConsentBuilder {
	f.consent.clientID = client.GetID()

	return f
}

func (f *FakeConsentBuilder) CreatedBy(user *users.User) *FakeConsentBuilder {
	f.consent.userID = user.ID()

	return f
}

func (f *FakeConsentBuilder) Build() *Consent {
	return f.consent
}

// func (f *FakeConsentBuilder) BuildAndStore(ctx context.Context, db *sql.DB) *Consent {
// 	f.t.Helper()
//
// 	tools := tools.NewToolboxForTest(f.t)
// 	storage := newSqlStorage(db, tools)
//
// 	err := storage.Save(ctx, f.consent)
// 	require.NoError(f.t, err)
//
// 	return f.consent
// }
