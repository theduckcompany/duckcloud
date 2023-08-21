package oauthclients

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var ClientIDRegexp = regexp.MustCompile("^[0-9a-zA-Z-]+$")

// Client client model
type Client struct {
	id             uuid.UUID
	name           string
	secret         string
	redirectURI    string
	userID         uuid.UUID
	createdAt      time.Time
	scopes         Scopes
	public         bool
	skipValidation bool
}

func (c *Client) SkipValidation() bool { return c.skipValidation }
func (c *Client) Name() string         { return c.name }
func (c *Client) RedirectURI() string  { return c.redirectURI }
func (c *Client) GetID() string        { return string(c.id) }
func (c *Client) GetSecret() string    { return c.secret }
func (c *Client) GetDomain() string    { return c.redirectURI }
func (c *Client) IsPublic() bool       { return c.public }
func (c *Client) GetUserID() string    { return string(c.userID) }

type CreateCmd struct {
	ID             uuid.UUID
	Name           string
	RedirectURI    string
	UserID         uuid.UUID
	Scopes         Scopes
	Public         bool
	SkipValidation bool
}

// Validate the fields.
func (cmd CreateCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.ID, v.Required, v.Length(3, 20), v.Match(ClientIDRegexp)),
		v.Field(&cmd.Name, v.Required, v.Length(3, 20), is.ASCII),
		v.Field(&cmd.RedirectURI, v.Required, is.URL),
		v.Field(&cmd.UserID, v.Required, is.UUIDv4),
		v.Field(&cmd.Scopes, v.Required, v.Required),
	)
}

type Scopes []string

func (t Scopes) String() string               { return strings.Join(t, ",") }
func (t Scopes) Value() (driver.Value, error) { return strings.Join(t, ","), nil }
func (t *Scopes) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("not a string")
	}

	*t = strings.Split(s, ",")

	return nil
}
