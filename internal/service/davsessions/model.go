package davsessions

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var DavSessionRegexp = regexp.MustCompile("^[0-9a-zA-Z- ]+$")

type DavSession struct {
	id        uuid.UUID
	userID    uuid.UUID
	name      string
	username  string
	password  string
	folders   Folders
	createdAt time.Time
}

func (u *DavSession) ID() uuid.UUID        { return u.id }
func (u *DavSession) UserID() uuid.UUID    { return u.userID }
func (u *DavSession) Name() string         { return u.name }
func (u *DavSession) Username() string     { return u.username }
func (u *DavSession) FoldersIDs() Folders  { return u.folders }
func (u *DavSession) CreatedAt() time.Time { return u.createdAt }

type Folders []uuid.UUID

func (t Folders) String() string {
	rawIDs := make([]string, len(t))

	for i, id := range t {
		rawIDs[i] = string(id)
	}

	return strings.Join(rawIDs, ",")
}
func (t Folders) Value() (driver.Value, error) { return t.String(), nil }
func (t *Folders) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("not a string")
	}

	rawIDs := strings.Split(s, ",")

	for _, id := range rawIDs {
		*t = append(*t, uuid.UUID(id))
	}

	return nil
}

type CreateCmd struct {
	Name    string
	UserID  uuid.UUID
	Folders []uuid.UUID
}

func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Name, v.Required, v.Match(DavSessionRegexp)),
		v.Field(&t.UserID, v.Required, is.UUIDv4),
		v.Field(&t.Folders, v.Required, v.Each(is.UUIDv4)),
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
