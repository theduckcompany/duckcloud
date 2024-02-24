package davsessions

import (
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var DavSessionRegexp = regexp.MustCompile("^[0-9a-zA-Z- ]+$")

type DavSession struct {
	createdAt time.Time
	id        uuid.UUID
	userID    uuid.UUID
	name      string
	username  string
	password  secret.Text
	spaceID   uuid.UUID
}

func (u *DavSession) ID() uuid.UUID        { return u.id }
func (u *DavSession) UserID() uuid.UUID    { return u.userID }
func (u DavSession) Name() string          { return u.name }
func (u *DavSession) Username() string     { return u.username }
func (u *DavSession) SpaceID() uuid.UUID   { return u.spaceID }
func (u *DavSession) CreatedAt() time.Time { return u.createdAt }

type CreateCmd struct {
	Name     string
	Username string
	UserID   uuid.UUID
	SpaceID  uuid.UUID
}

func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Name, v.Required, v.Match(DavSessionRegexp)),
		v.Field(&t.Username, v.Required, v.Length(1, 30)),
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.SpaceID, v.Required, is.UUIDv4),
	)
}

type DeleteCmd struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

func (t DeleteCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.SessionID, v.Required, is.UUIDv4),
	)
}
