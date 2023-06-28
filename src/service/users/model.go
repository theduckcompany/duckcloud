package users

import (
	"regexp"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var UsernameRegexp = regexp.MustCompile("^[0-9a-zA-Z-]+$")

// User representation
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	password  string    `json:"-"`
}

// CreateUserRequest represents an user creation request.
type CreateUserRequest struct {
	Username string
	Email    string
	Password string
}

// Validate the CreateUserRequest fields.
func (t CreateUserRequest) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Username, v.Required, v.Length(1, 20), v.Match(UsernameRegexp)),
		v.Field(&t.Email, v.Required, is.EmailFormat, v.Length(1, 1000)),
		v.Field(&t.Password, v.Required, v.Length(8, 200)),
	)
}
