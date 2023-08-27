package davsessions

import (
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var DavSessionRegexp = regexp.MustCompile("^[0-9a-zA-Z- ]+$")

type DavSession struct {
	id        uuid.UUID
	userID    uuid.UUID
	name      string
	username  string
	password  string
	fsRoot    uuid.UUID
	createdAt time.Time
}

func (u *DavSession) ID() uuid.UUID        { return u.id }
func (u *DavSession) UserID() uuid.UUID    { return u.userID }
func (u *DavSession) Name() string         { return u.name }
func (u *DavSession) Username() string     { return u.username }
func (u *DavSession) RootFS() uuid.UUID    { return u.fsRoot }
func (u *DavSession) CreatedAt() time.Time { return u.createdAt }

type CreateCmd struct {
	Name   string
	UserID uuid.UUID
	FSRoot uuid.UUID
}

func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Name, v.Required, v.Match(DavSessionRegexp)),
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.FSRoot, v.Required, is.UUIDv4),
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
