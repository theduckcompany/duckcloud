package users

import (
	"encoding/json"
	"regexp"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var UsernameRegexp = regexp.MustCompile("^[0-9a-zA-Z-]+$")

// User representation
type User struct {
	id        uuid.UUID
	username  string
	email     string
	createdAt time.Time
	fsRoot    uuid.UUID
	password  string
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":        u.id,
		"username":  u.username,
		"email":     u.email,
		"createdAt": u.createdAt,
		"fsRoot":    u.fsRoot,
	})
}

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) Username() string     { return u.username }
func (u *User) Email() string        { return u.email }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) RootFS() uuid.UUID    { return u.fsRoot }

// CreateCmd represents an user creation request.
type CreateCmd struct {
	Username string
	Email    string
	Password string
}

// Validate the CreateUserRequest fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Username, v.Required, v.Length(1, 20), v.Match(UsernameRegexp)),
		v.Field(&t.Email, v.Required, is.EmailFormat, v.Length(1, 1000)),
		v.Field(&t.Password, v.Required, v.Length(8, 200)),
	)
}
