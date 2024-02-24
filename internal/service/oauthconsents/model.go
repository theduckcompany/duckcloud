package oauthconsents

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Consent struct {
	createdAt    time.Time
	id           uuid.UUID
	userID       uuid.UUID
	sessionToken string
	clientID     string
	scopes       []string
}

func (c *Consent) ID() uuid.UUID        { return c.id }
func (c *Consent) UserID() uuid.UUID    { return c.userID }
func (c *Consent) SessionToken() string { return c.sessionToken }
func (c *Consent) ClientID() string     { return c.clientID }
func (c *Consent) Scopes() []string     { return c.scopes }
func (c *Consent) CreatedAt() time.Time { return c.createdAt }

type CreateCmd struct {
	UserID       uuid.UUID
	SessionToken string
	ClientID     string
	Scopes       []string
}

// Validate the fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.SessionToken, v.Required, is.UUIDv4),
		v.Field(&t.ClientID, v.Required, v.Match(oauthclients.ClientIDRegexp)),
		v.Field(&t.Scopes, v.Required, v.Length(1, 30)),
	)
}
