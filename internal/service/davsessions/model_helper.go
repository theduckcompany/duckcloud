package davsessions

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeSessionBuilder struct {
	session *DavSession
}

func NewFakeSession(t *testing.T) *FakeSessionBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()

	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeSessionBuilder{
		session: &DavSession{
			id:        uuidProvider.New(),
			name:      gofakeit.HipsterSentence(2),
			userID:    uuidProvider.New(),
			username:  gofakeit.Username(),
			password:  secret.NewText(gofakeit.Password(true, true, true, false, false, 8)),
			spaceID:   uuidProvider.New(),
			createdAt: createdAt,
		},
	}
}

func (f *FakeSessionBuilder) Build() *DavSession {
	return f.session
}
