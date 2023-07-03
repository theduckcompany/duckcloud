package oauthcodes

import (
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Code struct {
	Code            string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	ClientID        string
	UserID          string
	RedirectURI     string
	Scope           string
	Challenge       string
	ChallengeMethod string
}

type CreateCodeRequest struct {
	Code            string
	ExpiresAt       time.Time
	ClientID        string
	UserID          string
	RedirectURI     string
	Scope           string
	Challenge       string
	ChallengeMethod string
}

// Validate the fields.
func (t CreateCodeRequest) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Code, v.Required, v.Length(0, 200)),
		v.Field(&t.ExpiresAt, v.Required),
		v.Field(&t.ClientID, v.Length(3, 40), v.Match(regexp.MustCompile("^[0-9a-zA-Z-]+$"))),
		v.Field(&t.UserID, is.UUIDv4),
		v.Field(&t.RedirectURI, is.URL),
		v.Field(&t.Scope, v.Required),
		v.Field(&t.Challenge, v.Length(0, 100)),
		v.Field(&t.ChallengeMethod, v.In("plain", "S256")),
	)
}
