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
	token     string
	userID    uuid.UUID
	ip        string
	clientID  string
	device    string
	createdAt time.Time
}

func (s *Session) Token() string        { return s.token }
func (s *Session) UserID() uuid.UUID    { return s.userID }
func (s *Session) IP() string           { return s.ip }
func (s *Session) ClientID() string     { return s.clientID }
func (s *Session) Device() string       { return s.device }
func (s *Session) CreatedAt() time.Time { return s.createdAt }

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
