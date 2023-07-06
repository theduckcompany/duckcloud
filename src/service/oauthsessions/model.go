package oauthsessions

import (
	"regexp"
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

type CreateCmd struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
	ClientID         string
	UserID           string
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
