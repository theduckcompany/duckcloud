package oauthconsents

import (
	"time"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/tools/uuid"
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Consent struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	SessionToken string
	ClientID     string
	Scopes       []string
	CreatedAt    time.Time
}

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
