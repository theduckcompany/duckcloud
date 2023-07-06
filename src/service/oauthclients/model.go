package oauthclients

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// Client client model
type Client struct {
	ID             string
	Name           string
	Secret         string
	RedirectURI    string
	UserID         string
	CreatedAt      time.Time
	Scopes         Scopes
	Public         bool
	SkipValidation bool
}

type CreateCmd struct {
	ID             string
	Name           string
	RedirectURI    string
	UserID         string
	Scopes         Scopes
	Public         bool
	SkipValidation bool
}

// Validate the fields.
func (cmd CreateCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.ID, v.Length(3, 20), v.Match(regexp.MustCompile("^[0-9a-zA-Z-]+$"))),
		v.Field(&cmd.Name, v.Length(3, 20), is.ASCII),
		v.Field(&cmd.RedirectURI, is.URL),
		v.Field(&cmd.UserID, is.UUIDv4),
		v.Field(&cmd.Scopes, v.Required),
	)
}

// GetID client id
func (c *Client) GetID() string {
	return c.ID
}

// GetSecret client domain
func (c *Client) GetSecret() string {
	return c.Secret
}

// GetDomain client domain
func (c *Client) GetDomain() string {
	return c.RedirectURI
}

func (c *Client) IsPublic() bool {
	return c.Public
}

// GetUserID user id
func (c *Client) GetUserID() string {
	return c.UserID
}

type Scopes []string

func (t Scopes) String() string {
	return strings.Join(t, ",")
}

func (t Scopes) Value() (driver.Value, error) {
	return strings.Join(t, ","), nil
}

func (t *Scopes) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("not a string")
	}

	*t = strings.Split(s, ",")

	return nil
}
