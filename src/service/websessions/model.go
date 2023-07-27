package websessions

import (
	"net/http"
	"time"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/tools/uuid"
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Session struct {
	Token     string
	UserID    uuid.UUID
	IP        string
	ClientID  string
	Device    string
	CreatedAt time.Time
}

type CreateCmd struct {
	UserID   string
	ClientID string
	Req      *http.Request
}

// Validate the fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.ClientID, v.Required, v.Match(oauthclients.ClientIDRegexp)),
		v.Field(&t.Req, v.Required),
	)
}
