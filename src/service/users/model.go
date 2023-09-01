package users

import (
	"encoding/json"
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var UsernameRegexp = regexp.MustCompile("^[0-9a-zA-Z-]+$")

// User representation
type User struct {
	id        uuid.UUID
	username  string
	isAdmin   bool
	createdAt time.Time
	fsRoot    uuid.UUID
	password  string
	status    string
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":        u.id,
		"username":  u.username,
		"admin":     u.isAdmin,
		"createdAt": u.createdAt,
		"fsRoot":    u.fsRoot,
		"status":    u.status,
	})
}

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) Username() string     { return u.username }
func (u *User) IsAdmin() bool        { return u.isAdmin }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) RootFS() uuid.UUID    { return u.fsRoot }
func (u *User) Status() string       { return u.status }

// CreateCmd represents an user creation request.
type CreateCmd struct {
	Username string
	Password string
	IsAdmin  bool
}

// Validate the CreateUserRequest fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Username, v.Required, v.Length(1, 20), v.Match(UsernameRegexp)),
		v.Field(&t.Password, v.Required, v.Length(8, 200)),
		v.Field(&t.IsAdmin),
	)
}
