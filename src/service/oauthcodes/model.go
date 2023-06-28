package oauthcodes

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Code struct {
	Code        string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ClientID    string
	UserID      string
	RedirectURI string
	Scope       string
}

type CreateCodeRequest struct {
	Code        string
	ExpiresAt   time.Time
	ClientID    string
	UserID      string
	RedirectURI string
	Scope       string
}

// Validate the fields.
func (t CreateCodeRequest) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Code, v.Required),
		v.Field(&t.ExpiresAt, v.Required),
		v.Field(&t.ClientID, is.UUIDv4),
		v.Field(&t.UserID, is.UUIDv4),
		v.Field(&t.RedirectURI, is.URL),
		v.Field(&t.Scope, v.Required),
	)
}
