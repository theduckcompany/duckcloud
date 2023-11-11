package oauthcodes

import (
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

type Code struct {
	code            secret.Text
	createdAt       time.Time
	expiresAt       time.Time
	clientID        string
	userID          string
	redirectURI     string
	scope           string
	challenge       secret.Text
	challengeMethod string
}

func (c *Code) Code() secret.Text       { return c.code }
func (c *Code) CreatedAt() time.Time    { return c.createdAt }
func (c *Code) ExpiresAt() time.Time    { return c.expiresAt }
func (c *Code) ClientID() string        { return c.clientID }
func (c *Code) UserID() string          { return c.userID }
func (c *Code) RedirectURI() string     { return c.redirectURI }
func (c *Code) Scope() string           { return c.scope }
func (c *Code) Challenge() secret.Text  { return c.challenge }
func (c *Code) ChallengeMethod() string { return c.challengeMethod }

type CreateCmd struct {
	Code            secret.Text
	ExpiresAt       time.Time
	ClientID        string
	UserID          string
	RedirectURI     string
	Scope           string
	Challenge       secret.Text
	ChallengeMethod string
}

// Validate the fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Code, v.Required),
		v.Field(&t.ExpiresAt, v.Required),
		v.Field(&t.ClientID, v.Length(3, 40), v.Match(regexp.MustCompile("^[0-9a-zA-Z-]+$"))),
		v.Field(&t.UserID, is.UUIDv4),
		v.Field(&t.RedirectURI, is.URL),
		v.Field(&t.Scope, v.Required),
		v.Field(&t.Challenge),
		v.Field(&t.ChallengeMethod, v.In("plain", "S256")),
	)
}
