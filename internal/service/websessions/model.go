package websessions

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Session struct {
	createdAt time.Time
	token     secret.Text
	userID    uuid.UUID
	ip        string
	device    string
}

func (s *Session) Token() secret.Text   { return s.token }
func (s *Session) UserID() uuid.UUID    { return s.userID }
func (s *Session) IP() string           { return s.ip }
func (s *Session) Device() string       { return s.device }
func (s *Session) CreatedAt() time.Time { return s.createdAt }

type CreateCmd struct {
	UserID     uuid.UUID
	UserAgent  string
	RemoteAddr string
}

func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.UserAgent, v.Required),
		v.Field(&t.RemoteAddr, v.Required),
	)
}

type DeleteCmd struct {
	UserID uuid.UUID
	Token  secret.Text
}

func (t DeleteCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.Token, v.Required),
	)
}
