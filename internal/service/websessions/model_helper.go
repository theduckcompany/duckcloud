package websessions

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeSessionBuilder struct {
	session *Session
}

func NewFakeSession(t *testing.T) *FakeSessionBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()

	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())
	rawToken := gofakeit.Password(true, true, true, false, false, 8)

	return &FakeSessionBuilder{
		session: &Session{
			createdAt: createdAt,
			token:     secret.NewText(rawToken),
			userID:    uuidProvider.New(),
			ip:        gofakeit.IPv4Address(),
			device:    gofakeit.AppName(),
		},
	}
}

func (f *FakeSessionBuilder) CreatedAt(at time.Time) *FakeSessionBuilder {
	f.session.createdAt = at

	return f
}

func (f *FakeSessionBuilder) CreatedBy(user *users.User) *FakeSessionBuilder {
	f.session.userID = user.ID()

	return f
}

func (f *FakeSessionBuilder) WithToken(token string) *FakeSessionBuilder {
	f.session.token = secret.NewText(token)

	return f
}

func (f *FakeSessionBuilder) WithIP(ip string) *FakeSessionBuilder {
	f.session.ip = ip

	return f
}

func (f *FakeSessionBuilder) WithDevice(device string) *FakeSessionBuilder {
	f.session.device = device

	return f
}

func (f *FakeSessionBuilder) Build() *Session {
	return f.session
}
