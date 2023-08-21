package davsessions

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type DavSession struct {
	id        uuid.UUID
	userID    uuid.UUID
	username  string
	password  string
	fsRoot    uuid.UUID
	createdAt time.Time
}

func (u *DavSession) ID() uuid.UUID        { return u.id }
func (u *DavSession) UserID() uuid.UUID    { return u.userID }
func (u *DavSession) Username() string     { return u.username }
func (u *DavSession) RootFS() uuid.UUID    { return u.fsRoot }
func (u *DavSession) CreatedAt() time.Time { return u.createdAt }

type CreateCmd struct {
	UserID uuid.UUID
	FSRoot uuid.UUID
}

func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.FSRoot, v.Required, is.UUIDv4),
	)
}
