package websessions

import (
	"net/http"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
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

func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.ClientID, v.Required, v.Match(oauthclients.ClientIDRegexp)),
		v.Field(&t.Req, v.Required),
	)
}

type RevokeCmd struct {
	UserID uuid.UUID
	Token  string
}

func (t RevokeCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.Token, v.Required, is.UUIDv4),
	)
}
