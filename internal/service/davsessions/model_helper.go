package davsessions

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
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
	rawPassword := gofakeit.Password(true, true, true, false, false, 8)

	return &FakeSessionBuilder{
		session: &DavSession{
			createdAt: createdAt,
			id:        uuidProvider.New(),
			userID:    uuidProvider.New(),
			name:      gofakeit.Animal(),
			username:  gofakeit.Username(),
			password:  secret.NewText(hex.EncodeToString([]byte(rawPassword))),
			spaceID:   uuidProvider.New(),
		},
	}
}

func (f *FakeSessionBuilder) WithPassword(password string) *FakeSessionBuilder {
	f.session.password = secret.NewText(hex.EncodeToString([]byte(password)))

	return f
}

func (f *FakeSessionBuilder) WithUsername(username string) *FakeSessionBuilder {
	f.session.username = username

	return f
}

func (f *FakeSessionBuilder) WithName(name string) *FakeSessionBuilder {
	f.session.name = name

	return f
}

func (f *FakeSessionBuilder) WithSpace(space *spaces.Space) *FakeSessionBuilder {
	f.session.spaceID = space.ID()

	return f
}

func (f *FakeSessionBuilder) CreatedAt(at time.Time) *FakeSessionBuilder {
	f.session.createdAt = at

	return f
}

func (f *FakeSessionBuilder) CreatedBy(user *users.User) *FakeSessionBuilder {
	f.session.userID = user.ID()

	return f
}

func (f *FakeSessionBuilder) Build() *DavSession {
	return f.session
}
