package oauthsessions

import (
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Session struct {
	accessToken      secret.Text
	accessCreatedAt  time.Time
	accessExpiresAt  time.Time
	refreshToken     secret.Text
	refreshCreatedAt time.Time
	refreshExpiresAt time.Time
	clientID         string
	userID           uuid.UUID
	scope            string
}

func (s *Session) AccessToken() secret.Text    { return s.accessToken }
func (s *Session) AccessCreatedAt() time.Time  { return s.accessCreatedAt }
func (s *Session) AccessExpiresAt() time.Time  { return s.accessExpiresAt }
func (s *Session) RefreshToken() secret.Text   { return s.refreshToken }
func (s *Session) RefreshCreatedAt() time.Time { return s.refreshCreatedAt }
func (s *Session) RefreshExpiresAt() time.Time { return s.refreshExpiresAt }
func (s *Session) ClientID() string            { return s.clientID }
func (s *Session) UserID() uuid.UUID           { return s.userID }
func (s *Session) Scope() string               { return s.scope }

type CreateCmd struct {
	AccessToken      secret.Text
	AccessExpiresAt  time.Time
	RefreshToken     secret.Text
	RefreshExpiresAt time.Time
	ClientID         string
	UserID           uuid.UUID
	Scope            string
}

// Validate the fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.AccessToken, v.Required),
		v.Field(&t.AccessExpiresAt, v.Required),
		v.Field(&t.RefreshToken, v.Required),
		v.Field(&t.RefreshExpiresAt, v.Required),
		v.Field(&t.ClientID, v.Length(3, 40), v.Match(regexp.MustCompile("^[0-9a-zA-Z-]+$"))),
		v.Field(&t.UserID, is.UUIDv4),
	)
}
