package oauthsessions

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Session struct {
	AccessToken      string
	AccessCreatedAt  time.Time
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshCreatedAt time.Time
	RefreshExpiresAt time.Time
	ClientID         string
	UserID           string
	Scope            string
}

type CreateSessionRequest struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
	ClientID         string
	UserID           string
	Scope            string
}

// Validate the fields.
func (t CreateSessionRequest) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.AccessToken, v.Required),
		v.Field(&t.AccessExpiresAt, v.Required),
		v.Field(&t.RefreshToken, v.Required),
		v.Field(&t.RefreshExpiresAt, v.Required),
		v.Field(&t.ClientID, is.UUIDv4),
		v.Field(&t.UserID, is.UUIDv4),
	)
}
